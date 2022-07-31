package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXTableCachingPolicyTestResource struct{}

func TestAccADXTableCachingPolicy_basic(t *testing.T) {
	var entity TableCachingPolicy
	r := ADXTableCachingPolicyTestResource{}
	rtcBuilder := BuildResourceTestContext[TableCachingPolicy]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_table_caching_policy").
		DatabaseName("test-db").
		EntityType("caching").
		ReadStatementFunc(GetAccTestPolicyReadStatementFunc("table","caching")).
		IDParserFunc(GetAccTestPolicyIDParserFunc()).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				Config: r.basic(rtc),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "data_hot_span", "3d"),
				),
			},
		},
	})
}

func (this ADXTableCachingPolicyTestResource) basic(rtc *ResourceTestContext[TableCachingPolicy]) string {
	return fmt.Sprintf(`
	%s

	resource "%s" %s {
		database_name = "%s"
		table_name    = "${adx_table.test.name}"
		data_hot_span = "3d"
	}
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.DatabaseName)
}

func (this ADXTableCachingPolicyTestResource) template(rtc *ResourceTestContext[TableCachingPolicy]) string {
	return fmt.Sprintf(`
	resource "adx_table" "test" {
		database_name = "%s"
		name          = "%s"
		table_schema  = "f1:string,f2:string,f4:string,f3:int"
	}
	`, rtc.DatabaseName, rtc.EntityName)
}
