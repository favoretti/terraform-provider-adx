package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXFunctionTestResource struct{}

func TestAccADXFunction_basic(t *testing.T) {
	var entity ADXFunction
	r := ADXFunctionTestResource{}
	rtcBuilder := BuildResourceTestContext[ADXFunction]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_function").
		DatabaseName("test-db").
		EntityType("function").
		ReadStatementFunc(func(id string) (string, error) {
			funcId, err := parseADXFunctionID(id)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(".show functions | where Name == '%s'", funcId.Name), nil
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
					resource.TestCheckResourceAttr(rtc.GetTFName(), "parameters", "()"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "body", "{Test1 \n| limit 10}"),
				),
			},
			{
				Config: r.basic_update(rtc),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "parameters", "(param1:string, limitSize:int)"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "body", "{Test1 \n| limit limitSize}"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "docstring", "This is table"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "folder", "iamafolder"),
				),
			},
			{
				ResourceName:      rtc.GetTFName(),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func (this ADXFunctionTestResource) basic(rtc *ResourceTestContext[ADXFunction]) string {
	return fmt.Sprintf(`
	%s

	resource "%s" %s {
		database_name = "%s"
		name          = "%s"
		body          = "{${adx_table.test.name} \n| limit 10}"
	}
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName)
}

func (this ADXFunctionTestResource) basic_update(rtc *ResourceTestContext[ADXFunction]) string {
	return fmt.Sprintf(`
	%s

	resource "%s" %s {
		database_name = "%s"
		name          = "%s"
		body          = "{${adx_table.test.name} \n| limit limitSize}"
		parameters    = "(param1:string, limitSize:int)"
		docstring     = "This is table"
		folder        = "iamafolder"
	}
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName)
}

func (this ADXFunctionTestResource) template(rtc *ResourceTestContext[ADXFunction]) string {
	return fmt.Sprintf(`
	resource "adx_table" "test" {
		database_name = "%s"
		name          = "Test1"
		table_schema  = "f1:string,f2:string,f4:string,f3:int"
		docstring     = "This is table"
		folder        = "iamafolder"
	}
	`, rtc.DatabaseName)
}
