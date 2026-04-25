package adx

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type TablePrincipal struct {
	Role                 string
	PrincipalType        string
	PrincipalDisplayName string
	PrincipalObjectId    string
	PrincipalFQN         string
	Notes                string
}

func resourceADXTableSecurityRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableSecurityRoleCreate,
		ReadContext:   resourceADXTableSecurityRoleRead,
		UpdateContext: resourceADXTableSecurityRoleUpdate,
		DeleteContext: resourceADXTableSecurityRoleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"cluster": getClusterConfigInputSchema(),
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"table_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"admins",
					"ingestors",
				}, false),
			},

			"principal_fqn": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
				DiffSuppressFunc: suppressPrincipalFQNDiff,
				Description: "Fully qualified name of the principal, e.g. 'aaduser=user@example.com' or 'aadapp=<app-id>;<tenant>'",
			},

			"notes": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Free text notes describing the role assignment",
			},

			"principal_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the principal (e.g., AAD User, AAD App, AAD Group)",
			},

			"principal_display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Display name of the principal",
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXTableSecurityRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	role := d.Get("role").(string)
	principalFQN := d.Get("principal_fqn").(string)
	notes := d.Get("notes").(string)

	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}

	escapedTableName := escapeEntityNameIfRequired(tableName)

	addStatement := fmt.Sprintf(".add table %s %s ('%s')", escapedTableName, role, principalFQN)
	if notes != "" {
		addStatement = fmt.Sprintf("%s '%s'", addStatement, notes)
	}

	// The .add command returns the updated list of principals — parse it to get
	// the actual (resolved) FQN that ADX stores for this principal.
	resultSet, err := queryADXMgmtAndParse[TablePrincipal](ctx, meta, clusterConfig, databaseName, addStatement)
	if err != nil {
		return diag.Errorf("error adding %s role for principal %q on table %q (Database %q): %+v", role, principalFQN, tableName, databaseName, err)
	}

	displayRole := roleToDisplayName(role)
	var actualFQN string
	for _, p := range resultSet {
		if matchesPrincipal(p, principalFQN) && strings.HasSuffix(p.Role, displayRole) {
			actualFQN = p.PrincipalFQN
			break
		}
	}

	if actualFQN == "" {
		// Fallback: use the user-provided FQN if we can't resolve
		log.Printf("[WARN] Could not find resolved FQN for principal %q in .add response, using input value", principalFQN)
		actualFQN = principalFQN
	}

	id := buildADXResourceId(client.Endpoint(), databaseName, "table", tableName, "security_role", role, actualFQN)
	d.SetId(id)

	return resourceADXTableSecurityRoleRead(ctx, d, meta)
}

func resourceADXTableSecurityRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXTableSecurityRoleID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing resource ID: %+v", err)
	}

	if tableExists, err := isTableExists(ctx, meta, clusterConfig, id.DatabaseName, id.Name); err != nil || !tableExists {
		if err != nil {
			return diag.Errorf("%+v", err)
		}
		d.SetId("")
		return diags
	}

	escapedTableName := escapeEntityNameIfRequired(id.Name)
	showStatement := fmt.Sprintf(".show table %s principals", escapedTableName)

	resultSet, err := queryADXMgmtAndParse[TablePrincipal](ctx, meta, clusterConfig, id.DatabaseName, showStatement)
	if err != nil {
		return diag.Errorf("error reading principals for table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}

	displayRole := roleToDisplayName(id.Role)

	var found *TablePrincipal
	for _, p := range resultSet {
		if matchesPrincipal(p, id.PrincipalFQN) && strings.HasSuffix(p.Role, displayRole) {
			found = &p
			break
		}
	}

	if found == nil {
		d.SetId("")
		return diags
	}

	d.Set("database_name", id.DatabaseName)
	d.Set("table_name", id.Name)
	d.Set("role", id.Role)
	d.Set("principal_fqn", found.PrincipalFQN)
	d.Set("principal_type", found.PrincipalType)
	d.Set("principal_display_name", found.PrincipalDisplayName)
	d.Set("notes", found.Notes)

	return diags
}

func resourceADXTableSecurityRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXTableSecurityRoleID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing resource ID: %+v", err)
	}

	notes := d.Get("notes").(string)
	escapedTableName := escapeEntityNameIfRequired(id.Name)

	dropStatement := fmt.Sprintf(".drop table %s %s ('%s')", escapedTableName, id.Role, id.PrincipalFQN)
	_, err = queryADXMgmt(ctx, meta, clusterConfig, id.DatabaseName, dropStatement)
	if err != nil {
		return diag.Errorf("error dropping %s role for principal %q on table %q (Database %q) during update: %+v", id.Role, id.PrincipalFQN, id.Name, id.DatabaseName, err)
	}

	addStatement := fmt.Sprintf(".add table %s %s ('%s')", escapedTableName, id.Role, id.PrincipalFQN)
	if notes != "" {
		addStatement = fmt.Sprintf("%s '%s'", addStatement, notes)
	}

	_, err = queryADXMgmt(ctx, meta, clusterConfig, id.DatabaseName, addStatement)
	if err != nil {
		return diag.Errorf("error re-adding %s role for principal %q on table %q (Database %q) during update: %+v", id.Role, id.PrincipalFQN, id.Name, id.DatabaseName, err)
	}

	return resourceADXTableSecurityRoleRead(ctx, d, meta)
}

func resourceADXTableSecurityRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXTableSecurityRoleID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing resource ID: %+v", err)
	}

	escapedTableName := escapeEntityNameIfRequired(id.Name)
	dropStatement := fmt.Sprintf(".drop table %s %s ('%s')", escapedTableName, id.Role, id.PrincipalFQN)

	return deleteADXEntity(ctx, d, meta, clusterConfig, id.DatabaseName, dropStatement)
}

type adxTableSecurityRoleResourceId struct {
	adxResourceId
	Role         string
	PrincipalFQN string
}

func parseADXTableSecurityRoleID(input string) (*adxTableSecurityRoleResourceId, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 7 {
		return nil, fmt.Errorf("error parsing ADX Table Security Role resource ID: unexpected format: %q, expected 7 parts separated by '|'", input)
	}

	return &adxTableSecurityRoleResourceId{
		adxResourceId: adxResourceId{
			EndpointURI:  parts[0],
			DatabaseName: parts[1],
			EntityType:   parts[2],
			Name:         parts[3],
		},
		// parts[4] = "security_role"
		Role:         parts[5],
		PrincipalFQN: parts[6],
	}, nil
}

// roleToDisplayName maps KQL command role names (plural) to the display names
// returned by `.show table <T> principals` (singular).
func roleToDisplayName(role string) string {
	switch role {
	case "admins":
		return "Admin"
	case "ingestors":
		return "Ingestor"
	default:
		return role
	}
}

// matchesPrincipal checks if an ADX principal matches the user-provided FQN.
// ADX resolves e.g. "aaduser=user@example.com" to "aaduser=<guid>;<tenant-guid>",
// so we must also check the PrincipalDisplayName which contains "(upn: user@example.com)".
func matchesPrincipal(p TablePrincipal, inputFQN string) bool {
	if inputFQN == "" {
		return false
	}
	actualLower := strings.ToLower(p.PrincipalFQN)
	inputLower := strings.ToLower(inputFQN)

	// Exact match
	if actualLower == inputLower {
		return true
	}

	// Substring match on FQN (works for app IDs and GUIDs)
	if strings.Contains(actualLower, inputLower) || strings.Contains(inputLower, actualLower) {
		return true
	}

	// Extract the identifier after '=' (e.g., "aaduser=user@example.com" → "user@example.com")
	// and check if it appears in PrincipalDisplayName (which contains the UPN for resolved users)
	if parts := strings.SplitN(inputFQN, "=", 2); len(parts) == 2 {
		identifier := strings.ToLower(parts[1])
		if identifier != "" && p.PrincipalDisplayName != "" {
			if strings.Contains(strings.ToLower(p.PrincipalDisplayName), identifier) {
				return true
			}
		}
	}

	return false
}

// suppressPrincipalFQNDiff suppresses diffs between user-provided FQN and ADX-resolved FQN.
// After Read stores the resolved FQN (e.g. "aaduser=<guid>;<tenant>"), the next plan would
// see a diff vs the config value (e.g. "aaduser=user@example.com"). We suppress this by
// checking that both share the same principal type prefix.
func suppressPrincipalFQNDiff(k, old, new string, d *schema.ResourceData) bool {
	if old == "" || new == "" {
		return old == new
	}
	if strings.EqualFold(old, new) {
		return true
	}
	// Both must have the same principal type prefix (aaduser, aadapp, aadgroup, etc.)
	oldParts := strings.SplitN(strings.ToLower(old), "=", 2)
	newParts := strings.SplitN(strings.ToLower(new), "=", 2)
	if len(oldParts) == 2 && len(newParts) == 2 && oldParts[0] == newParts[0] {
		// Same principal type - suppress the diff since ADX resolves FQNs
		return true
	}
	return false
}
