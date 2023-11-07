package adx

import (
	"context"
	"fmt"
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
				Type:     schema.TypeString,
				Optional: true,
				Default:  "jsonencode([])",
			},

			"path_format": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder": {
				Type:     schema.TypeString,
				Optional: true,
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
	dataFormat := d.Get("data_format").(string)
	storageConnectionString := d.Get("storage_connection_string").(string)
	schema := d.Get("schema").(string)
	partitions := d.Get("partitions").(string)
	pathFormat := d.Get("path_format").(string)
	kind := d.Get("kind").(string)

	var withParams []string

	if folder, ok := d.GetOk("folder"); ok {
		withParams = append(withParams, fmt.Sprintf("folder=%s", folder.(string)))
	}
	if docString, ok := d.GetOk("doc_string"); ok {
		withParams = append(withParams, fmt.Sprintf("docString=%s", docString.(string)))
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

	createStatement := ""
	if len(partitions) > 0 && len(pathFormat) > 0 {
		createStatement = fmt.Sprintf(".create-or-alter external table %s (%s) kind = %s partition by (%s) pathformat = (%s) dataformat = %s ('%s') %s", name, schema, kind, partitions, pathFormat, dataFormat, storageConnectionString, withClause)
	} else if len(partitions) > 0 && len(pathFormat) == 0 {
		createStatement = fmt.Sprintf(".create-or-alter external table %s (%s) kind = %s partition by (%s) dataformat = %s ('%s') %s", name, schema, kind, partitions, dataFormat, storageConnectionString, withClause)
	} else if len(partitions) == 0 && len(pathFormat) > 0 {
		createStatement = fmt.Sprintf(".create-or-alter external table %s (%s) kind = %s pathformat = (%s) dataformat = %s ('%s') %s", name, schema, kind, pathFormat, dataFormat, storageConnectionString, withClause)
	} else {
		createStatement = fmt.Sprintf(".create-or-alter external table %s (%s) kind = %s dataformat = %s ('%s') %s", name, schema, kind, dataFormat, storageConnectionString, withClause)
	}

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
		d.Set("storage_connection_string", resultSet[0].ConnectionStrings)
		d.Set("partitions", resultSet[0].Partitions)
		d.Set("path_format", resultSet[0].PathFormat)
		d.Set("folder", resultSet[0].Folder)
		d.Set("properties", resultSet[0].Properties)

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
