package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXTableTestResource struct{}

func TestAccADXTable_basic(t *testing.T) {
	var entity TableSchema
	r := ADXTableTestResource{}
	rtcBuilder := BuildResourceTestContext[TableSchema]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_table").
		DatabaseName("test-db").
		EntityType("table").
		ReadStatementFunc(func(id string) (string, error) {
			funcId, err := parseADXTableID(id)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(".show tables | where TableName == '%s'", funcId.Name), nil
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
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_schema", "f1:string,f2:string,f4:string,f3:int"),
				),
			},
			{
				Config: r.basic_update(rtc),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_schema", "f1:string,f2:string,f4:string,f3:int"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "docstring", "This is table"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "folder", "iamafolder"),
				),
			},
		},
	})
}

func (this ADXTableTestResource) basic(rtc *ResourceTestContext[TableSchema]) string {
	return fmt.Sprintf(`

	resource "%s" %s {
		database_name = "%s"
		name          = "%s"
		table_schema  = "f1:string,f2:string,f4:string,f3:int"
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName)
}

func (this ADXTableTestResource) basic_update(rtc *ResourceTestContext[TableSchema]) string {
	return fmt.Sprintf(`

	resource "%s" %s {
		database_name = "%s"
		name          = "%s"
		table_schema  = "f1:string,f2:string,f4:string,f3:int"
		docstring     = "This is table"
		folder        = "iamafolder"
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName)
}
