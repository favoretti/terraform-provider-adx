package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXMaterializedViewTestResource struct{}

func TestAccMaterializedView(t *testing.T) {
	var entity ADXMaterializedView
	tableName := "MvTest1"
	r := ADXMaterializedViewTestResource{}
	rtcBuilder := BuildResourceTestContext[ADXMaterializedView]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_materialized_view").
		DatabaseName("test-db").
		EntityType("materializedview").
		ReadStatementFunc(func(id string) (string, error) {
			viewId, err := parseADXMaterializedViewID(id)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(".show materialized-views | where Name == '%s' | extend Lookback=tostring(Lookback), IsHealthy=tolower(tostring(IsHealthy)), IsEnabled=tolower(tostring(IsEnabled)), AutoUpdateSchema=tolower(tostring(AutoUpdateSchema)), EffectiveDateTime", viewId.Name), nil
		}).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				ImportState:       true,
				ImportStateVerify: true,
				Config:            r.basicMv(rtc, tableName, ""),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "source_table_name", tableName),
					rtc.CheckQueryResultSize(rtc.EntityName, 6, "Materialized view query check"),
				),
			},
			{
				Config: r.basicMv(rtc, tableName, "| extend newcol = 1"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "source_table_name", tableName),
					rtc.CheckQueryResultSize(rtc.EntityName, 6, "Materialized view query check"),
				),
			},
		},
	})
}

func TestAccMaterializedView_RLSSourceTable(t *testing.T) {
	var entity ADXMaterializedView
	tableName := "MvTest1RLS"
	r := ADXMaterializedViewTestResource{}
	rtcBuilder := BuildResourceTestContext[ADXMaterializedView]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_materialized_view").
		DatabaseName("test-db").
		EntityType("materializedview").
		ReadStatementFunc(func(id string) (string, error) {
			viewId, err := parseADXMaterializedViewID(id)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(".show materialized-views | where Name == '%s' | extend Lookback=tostring(Lookback), IsHealthy=tolower(tostring(IsHealthy)), IsEnabled=tolower(tostring(IsEnabled)), AutoUpdateSchema=tolower(tostring(AutoUpdateSchema)), EffectiveDateTime", viewId.Name), nil
		}).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				Config: r.mvRLSTable(rtc, tableName, ""),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "source_table_name", tableName),
					rtc.CheckQueryResultSize(rtc.EntityName, 6, "Materialized view query check"),
				),
			},
			{
				Config: r.mvRLSTable(rtc, tableName, "| extend newcol = 1"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "source_table_name", tableName),
					rtc.CheckQueryResultSize(rtc.EntityName, 6, "Materialized view query check"),
				),
			},
		},
	})
}

func (this ADXMaterializedViewTestResource) basicMv(rtc *ResourceTestContext[ADXMaterializedView], tableName string, extraClause string) string {
	return fmt.Sprintf(`
	%s

	resource "%s" "%s" {
		name              = "%s"
		database_name     = "%s"
		source_table_name = adx_table.%s.name
		backfill          = true
		query             = "${adx_table.%s.name} %s | summarize arg_max(score,*) by team"
	  }
	`, this.basicTable(rtc, tableName), rtc.Type, rtc.Label, rtc.EntityName, rtc.DatabaseName, rtc.Label, rtc.Label, extraClause)
}

func (this ADXMaterializedViewTestResource) mvRLSTable(rtc *ResourceTestContext[ADXMaterializedView], tableName string, extraClause string) string {
	return fmt.Sprintf(`
	%s

	resource "%s" "%s" {
		name                 = "%s"
		database_name        = "%s"
		source_table_name    = adx_table.%s.name
		backfill             = true
		query                = "${adx_table.%s.name} %s | summarize arg_max(score,*) by team"
		allow_mv_without_rls = true
	  }
	`, this.rlsTable(rtc, tableName), rtc.Type, rtc.Label, rtc.EntityName, rtc.DatabaseName, rtc.Label, rtc.Label, extraClause)
}

func (this ADXMaterializedViewTestResource) basicTable(rtc *ResourceTestContext[ADXMaterializedView], tableName string) string {
	return fmt.Sprintf(`
	resource "adx_table" "%s" {
		database_name    = "%s"
		name             = "%s"
		from_query {
			query = <<EOT
				let T = datatable(team:string, year: string, score:int)
				[
					"wildcats","1996",89,
					"wildcats","1997",44,
					"bears","1996",34,
					"bears","1997",77,
					"eagles","1996",65,
					"eagles","1997",62,
					"lizards","1996",96,
					"lizards","1997",56,
					"tigers","1997",20,
					"tigers","1996",90,
					"lions","1996",81,
					"lions","1997",34
				];
				T
			EOT
			append = false
		}
	}
	`, rtc.Label, rtc.DatabaseName, tableName)
}

func (this ADXMaterializedViewTestResource) rlsTable(rtc *ResourceTestContext[ADXMaterializedView], tableName string) string {
	return fmt.Sprintf(`
	%s

	resource "adx_function" "%s" {
		database_name = "%s"
		name          = "test_rls_function_mv"
		body          = "{${adx_table.%s.name} | where year == '1996'}"
	}
	  
	resource "adx_table_row_level_security_policy" "%s" {
		database_name = "%s"
		table_name    = adx_table.%s.name
		query         = adx_function.%s.name
		enabled       = true

		allow_mv_without_rls = true
	}
	`, this.basicTable(rtc, tableName), rtc.Label, rtc.DatabaseName, rtc.Label, rtc.Label, rtc.DatabaseName, rtc.Label, rtc.Label)
}
