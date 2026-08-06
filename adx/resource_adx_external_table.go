package adx

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ADXExternalTable struct {
	Name              string
	ConnectionStrings string
	Partitions        string
	PathFormat        string
	Folder            string
	Properties        string
}

func resourceADXExternalTable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXExternalTableCreateUpdate,
		UpdateContext: resourceADXExternalTableCreateUpdate,
		ReadContext:   resourceADXExternalTableRead,
		DeleteContext: resourceADXExternalTableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cluster": getClusterConfigInputSchema(),
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"data_format": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"storage_connection_string": {
				Type:             schema.TypeString,
				Required:         true,
				Sensitive:        true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"schema": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"kind": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "storage",
			},

			"partitions": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "jsonencode([])",
				DiffSuppressFunc: suppressExternalTablePartitionsDiff,
			},

			"path_format": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressExternalTablePathFormatDiff,
			},

			"folder": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressExternalTableFolderDiff,
			},

			"doc_string": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"compressed": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"include_headers": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "All",
			},

			"name_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"file_extension": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"encoding": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "UTF8NoBOM",
			},

			"sample_uris": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"files_preview": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"validate_not_empty": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"dry_run": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceADXExternalTableCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
	name := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	createStatement := buildExternalTableCreateStatement(d)

	_, err := queryADXMgmt(ctx, meta, clusterConfig, databaseName, createStatement)
	if err != nil {
		return diag.Errorf("error creating external table %s (Database %q): %+v", name, databaseName, err)
	}

	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}
	d.SetId(buildADXResourceId(client.Endpoint(), databaseName, "externaltable", name))

	return resourceADXExternalTableRead(ctx, d, meta)
}

// buildExternalTableCreateStatement assembles the .create-or-alter KQL statement sent to
// ADX for both create and update. partitions, path_format and folder are read via
// configuredExternalTableString rather than d.Get so the statement always reflects the
// KQL the user configured, not a stale ADX-canonicalized value suppressed onto state.
func buildExternalTableCreateStatement(d *schema.ResourceData) string {
	name := d.Get("name").(string)
	dataFormat := d.Get("data_format").(string)
	storageConnectionString := d.Get("storage_connection_string").(string)
	tableSchema := d.Get("schema").(string)
	partitions := configuredExternalTableString(d, "partitions")
	pathFormat := configuredExternalTableString(d, "path_format")
	kind := d.Get("kind").(string)

	var withParams []string

	if folder := configuredExternalTableString(d, "folder"); folder != "" {
		withParams = append(withParams, fmt.Sprintf("folder=%s", folder))
	}
	if docString, ok := d.GetOk("doc_string"); ok {
		withParams = append(withParams, fmt.Sprintf("docString='%s'", docString.(string)))
	}
	if compressed, ok := d.GetOk("compressed"); ok {
		withParams = append(withParams, fmt.Sprintf("compressed=%t", compressed.(bool)))
	}
	if includeHeaders, ok := d.GetOk("include_headers"); ok {
		withParams = append(withParams, fmt.Sprintf("includeHeaders=%s", includeHeaders.(string)))
	}
	if namePrefix, ok := d.GetOk("name_prefix"); ok {
		withParams = append(withParams, fmt.Sprintf("namePrefix=%s", namePrefix.(string)))
	}
	if fileExtension, ok := d.GetOk("file_extension"); ok {
		withParams = append(withParams, fmt.Sprintf("fileExtension='%s'", fileExtension.(string)))
	}
	if encoding, ok := d.GetOk("encoding"); ok {
		withParams = append(withParams, fmt.Sprintf("encoding='%s'", encoding.(string)))
	}
	if sampleUris, ok := d.GetOk("sample_uris"); ok {
		withParams = append(withParams, fmt.Sprintf("sampleUris=%t", sampleUris.(bool)))
	}
	if filesPreview, ok := d.GetOk("files_preview"); ok {
		withParams = append(withParams, fmt.Sprintf("filesPreview=%t", filesPreview.(bool)))
	}
	if validateNotEmpty, ok := d.GetOk("validate_not_empty"); ok {
		withParams = append(withParams, fmt.Sprintf("validateNotEmpty=%t", validateNotEmpty.(bool)))
	}
	if dryRun, ok := d.GetOk("dry_run"); ok {
		withParams = append(withParams, fmt.Sprintf("dryRun=%t", dryRun.(bool)))
	}

	withClause := ""
	if len(withParams) > 0 {
		withClause = fmt.Sprintf("with(%s)", strings.Join(withParams, ", "))
	}

	if len(partitions) > 0 && len(pathFormat) > 0 {
		return fmt.Sprintf(".create-or-alter external table %s (%s) kind = %s partition by (%s) pathformat = (%s) dataformat = %s ('%s') %s", name, tableSchema, kind, partitions, pathFormat, dataFormat, storageConnectionString, withClause)
	} else if len(partitions) > 0 && len(pathFormat) == 0 {
		return fmt.Sprintf(".create-or-alter external table %s (%s) kind = %s partition by (%s) dataformat = %s ('%s') %s", name, tableSchema, kind, partitions, dataFormat, storageConnectionString, withClause)
	} else if len(partitions) == 0 && len(pathFormat) > 0 {
		return fmt.Sprintf(".create-or-alter external table %s (%s) kind = %s pathformat = (%s) dataformat = %s ('%s') %s", name, tableSchema, kind, pathFormat, dataFormat, storageConnectionString, withClause)
	}
	return fmt.Sprintf(".create-or-alter external table %s (%s) kind = %s dataformat = %s ('%s') %s", name, tableSchema, kind, dataFormat, storageConnectionString, withClause)
}

func resourceADXExternalTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXExternalTableID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	showCommand := fmt.Sprintf(".show external table %s ", id.Name)

	resultSet, diags := readADXEntity[ADXExternalTable](ctx, meta, clusterConfig, id, showCommand, "external table")
	if diags.HasError() {
		return diags
	}

	if len(resultSet) < 1 {
		d.SetId("")
	} else {

		d.Set("name", id.Name)
		d.Set("database_name", id.DatabaseName)

		// ADX never returns the connection string verbatim (secret is redacted), so keep
		// the configured value and only read it when state is empty (e.g. on import).
		if v, ok := d.GetOk("storage_connection_string"); !ok || v.(string) == "" {
			d.Set("storage_connection_string", resultSet[0].ConnectionStrings)
		}

		d.Set("partitions", resultSet[0].Partitions)
		d.Set("path_format", resultSet[0].PathFormat)
		d.Set("folder", resultSet[0].Folder)

	}

	return diags
}

func resourceADXExternalTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXExternalTableID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return deleteADXEntity(ctx, d, meta, clusterConfig, id.DatabaseName, fmt.Sprintf(".drop external table %s", id.Name))
}

func parseADXExternalTableID(input string) (*adxResourceId, error) {
	return parseADXResourceID(input, 4, 0, 1, 2, 3)
}

// configuredExternalTableString returns the value the user actually wrote for key in
// their current configuration, bypassing d.Get. Once state holds ADX's canonicalized
// form (e.g. JSON partitions) and a DiffSuppressFunc below considers it equivalent to
// the KQL config, the SDK drops that attribute from the diff entirely, so d.Get would
// return the stale canonicalized state value instead of the KQL to send to ADX. Falls
// back to d.Get when the key is absent from config (e.g. relying on its schema default).
func configuredExternalTableString(d *schema.ResourceData, key string) string {
	raw := d.GetRawConfig()
	if raw.IsNull() {
		return d.Get(key).(string)
	}
	v := raw.GetAttr(key)
	if v.IsNull() || !v.IsKnown() {
		return d.Get(key).(string)
	}
	return v.AsString()
}

// ADX returns folder, path_format and partitions in a canonicalized form that differs
// from the KQL config syntax. These DiffSuppressFuncs compare them semantically.

// folder: config is quoted, ADX returns it unquoted.
func suppressExternalTableFolderDiff(k, old, new string, d *schema.ResourceData) bool {
	return normalizeExternalTableFolder(old) == normalizeExternalTableFolder(new)
}

func normalizeExternalTableFolder(s string) string {
	return strings.Trim(strings.TrimSpace(s), "'\"")
}

// path_format: datetime_pattern('yyyy/MM/dd', Day) vs datetime_pattern("yyyy/MM/dd",Day).
func suppressExternalTablePathFormatDiff(k, old, new string, d *schema.ResourceData) bool {
	return normalizeExternalTablePathFormat(old) == normalizeExternalTablePathFormat(new)
}

func normalizeExternalTablePathFormat(s string) string {
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, " ", "")
	return strings.TrimSpace(s)
}

// partitions: KQL clause (config) vs JSON array (ADX).
func suppressExternalTablePartitionsDiff(k, old, new string, d *schema.ResourceData) bool {
	oldCanon, oldOk := canonicalizeExternalTablePartitions(old)
	newCanon, newOk := canonicalizeExternalTablePartitions(new)

	// Fall back to literal comparison if either side is unparseable.
	if !oldOk || !newOk {
		return old == new
	}
	return oldCanon == newCanon
}

type externalTablePartition struct {
	Name     string
	Column   string
	Function string
	Mod      string
	Ordinal  int
}

// canonicalizeExternalTablePartitions reduces a KQL or JSON partition definition to a
// comparable signature. The bool reports whether parsing succeeded.
func canonicalizeExternalTablePartitions(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" || s == "[]" || s == "jsonencode([])" {
		return "", true
	}

	var parts []externalTablePartition
	var ok bool
	if strings.HasPrefix(s, "[") {
		parts, ok = parseExternalTablePartitionsJSON(s)
	} else {
		parts, ok = parseExternalTablePartitionsKQL(s)
	}
	if !ok {
		return "", false
	}

	var b strings.Builder
	for _, p := range parts {
		b.WriteString(strings.ToLower(p.Name))
		b.WriteByte('=')
		b.WriteString(strings.ToLower(p.Function))
		b.WriteByte('(')
		b.WriteString(strings.ToLower(p.Column))
		if p.Mod != "" {
			b.WriteByte(',')
			b.WriteString(p.Mod)
		}
		b.WriteString(");")
	}
	return b.String(), true
}

func parseExternalTablePartitionsJSON(s string) ([]externalTablePartition, bool) {
	var raw []struct {
		Name       string      `json:"Name"`
		ColumnName string      `json:"ColumnName"`
		Function   string      `json:"Function"`
		Mod        json.Number `json:"Mod"`
		Ordinal    int         `json:"Ordinal"`
	}
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return nil, false
	}

	sort.SliceStable(raw, func(i, j int) bool { return raw[i].Ordinal < raw[j].Ordinal })

	parts := make([]externalTablePartition, 0, len(raw))
	for _, r := range raw {
		parts = append(parts, externalTablePartition{
			Name:     r.Name,
			Column:   r.ColumnName,
			Function: r.Function,
			Mod:      strings.TrimSpace(r.Mod.String()),
			Ordinal:  r.Ordinal,
		})
	}
	return parts, true
}

var (
	kqlPartitionRe = regexp.MustCompile(`^([^\s:]+)\s*:\s*\w+\s*(?:=\s*(.+))?$`)
	kqlFunctionRe  = regexp.MustCompile(`^(\w+)\s*\((.+)\)$`)
)

func parseExternalTablePartitionsKQL(s string) ([]externalTablePartition, bool) {
	elements := splitTopLevel(s, ',')
	parts := make([]externalTablePartition, 0, len(elements))
	for _, el := range elements {
		el = strings.TrimSpace(el)
		if el == "" {
			continue
		}

		m := kqlPartitionRe.FindStringSubmatch(el)
		if m == nil {
			return nil, false
		}

		p := externalTablePartition{Name: strings.TrimSpace(m[1])}
		expr := strings.TrimSpace(m[2])
		if expr != "" {
			if fm := kqlFunctionRe.FindStringSubmatch(expr); fm != nil {
				p.Function = strings.TrimSpace(fm[1])
				args := splitTopLevel(fm[2], ',')
				if len(args) > 0 {
					p.Column = strings.TrimSpace(args[0])
				}
				if len(args) > 1 {
					p.Mod = strings.TrimSpace(args[1])
				}
			} else {
				p.Column = expr
			}
		}
		parts = append(parts, p)
	}
	return parts, true
}

// splitTopLevel splits s on sep, ignoring separators nested inside parentheses.
func splitTopLevel(s string, sep byte) []string {
	var res []string
	depth := 0
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case sep:
			if depth == 0 {
				res = append(res, s[start:i])
				start = i + 1
			}
		}
	}
	res = append(res, s[start:])
	return res
}
