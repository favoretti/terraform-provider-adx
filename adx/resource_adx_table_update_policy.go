package adx

import (
	"context"
	"fmt"

	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type TableUpdatePolicy struct {
	IsEnabled bool
	Source string
	Query string
	IsTransactional bool
	PropagateIngestionProperties bool
}

func resourceADXTableUpdatePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableUpdatePolicyCreate,
		ReadContext:   resourceADXTableUpdatePolicyRead,
		DeleteContext: resourceADXTableUpdatePolicyDelete,
		UpdateContext: resourceADXTableUpdatePolicyCreate,

		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"table_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew: 		  true,
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"enabled": {
				Type:             schema.TypeBool,
				Optional: true,
				Default: true,
			},

			"source_table": {
				Type:             schema.TypeString,
				Required: true,
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"query": {
				Type:             schema.TypeString,
				Required: true,
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"transactional": {
				Type:             schema.TypeBool,
				Required: true,
			},

			"propagate_ingestion_properties": {
				Type:             schema.TypeBool,
				Optional: true,
				Default: false,
			},
		},
	}
}

func resourceADXTableUpdatePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	enabled := d.Get("enabled").(bool)
	sourceTable := d.Get("source_table").(string)
	query := d.Get("query").(string)
	transactional := d.Get("transactional").(bool)
	propagateIngestionProperties := d.Get("propagate_ingestion_properties").(bool)

	createStatement := fmt.Sprintf(".alter table %s policy update @'[{\"IsEnabled\": %t, \"Source\": \"%s\", \"Query\": \"%s\", \"IsTransactional\": %t, \"PropagateIngestionProperties\": %t}]'", tableName, enabled, sourceTable, query, transactional, propagateIngestionProperties)

	if err := createADXPolicy(ctx, d, meta, "table","update", databaseName, tableName, createStatement); err != nil {
		return err
	}

	return resourceADXTableUpdatePolicyRead(ctx, d, meta)
}

func resourceADXTableUpdatePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	err, id, resultSet:= readADXPolicy(ctx,d,meta,"table","update"); 
	if err != nil {
		return diag.Errorf("%+v", err)
	}

	var policy []TableUpdatePolicy
	if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
		return diag.Errorf("error parsing policy update for Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}

	d.Set("table_name", id.Name)
	d.Set("database_name", id.DatabaseName)
	d.Set("enabled", policy[0].IsEnabled)
	d.Set("source_table", policy[0].Source)
	d.Set("query", policy[0].Query)
	d.Set("transactional", policy[0].IsTransactional)
	d.Set("propagate_ingestion_properties", policy[0].PropagateIngestionProperties)

	return diags
}

func resourceADXTableUpdatePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "table", "update")
}