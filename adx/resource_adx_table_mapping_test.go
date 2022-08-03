package adx

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXTableMappingTestResource struct{}

func TestAccTableMapping_basic(t *testing.T) {
	var entity TableMapping
	r := ADXTableMappingTestResource{}
	rtcBuilder := BuildResourceTestContext[TableMapping]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_table_mapping").
		DatabaseName("test-db").
		EntityType("tablemapping").
		ReadStatementFunc(func(id string) (string, error) {
			mappingId, err := parseADXTableMappingID(id)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(".show table %s ingestion %s mapping '%s'", mappingId.Name, strings.ToLower(mappingId.Kind), mappingId.MappingName), nil
		}).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				Config: r.basic(rtc),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", "MappingTest1"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "kind", "json"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "mapping.0.column", "f1"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "mapping.1.path", "$.something2.subtype"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "mapping.2.datatype", "int"),
				),
			},
		},
	})
}

func (this ADXTableMappingTestResource) basic(rtc *ResourceTestContext[TableMapping]) string {
	return fmt.Sprintf(`
	%s

	resource "%s" "%s" {
		name          = "%s"
		database_name = "%s"
		table_name    = adx_table.%s.name
		kind          = "json"
		mapping {
		  column   = "f1"
		  path     = "$.something1"
		  datatype = "string"
		}
		mapping {
		  column   = "f2"
		  path     = "$.something2.subtype"
		  datatype = "string"
		}
		mapping {
		  column   = "f3"
	      path     = "$.something3"
		  datatype = "int"
		}
	  }
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.EntityName, rtc.DatabaseName, rtc.Label)
}

func (this ADXTableMappingTestResource) template(rtc *ResourceTestContext[TableMapping]) string {
	return fmt.Sprintf(`
	resource "adx_table" "%s" {
		database_name = "%s"
		name          = "MappingTest1"
		table_schema  = "f1:string,f2:string,f3:int"
	}
	`, rtc.Label, rtc.DatabaseName)
}
