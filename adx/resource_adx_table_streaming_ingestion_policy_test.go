package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXTableStreamingIngestionPolicyTestResource struct{}

func TestAccADXTableStreamingIngestionPolicy_basic(t *testing.T) {
	var entity TableStreamingIngestionPolicy
	r := ADXTableStreamingIngestionPolicyTestResource{}
	rtcBuilder := BuildResourceTestContext[TableStreamingIngestionPolicy]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_table_streaming_ingestion_policy").
		DatabaseName("test-db").
		EntityType("streamingingestion").
		ReadStatementFunc(GetAccTestPolicyReadStatementFunc()).Build()

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
					resource.TestCheckResourceAttr(rtc.GetTFName(), "enabled", "true"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "hint_allocated_rate", "2.1"),
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

func (this ADXTableStreamingIngestionPolicyTestResource) basic(rtc *ResourceTestContext[TableStreamingIngestionPolicy]) string {
	return fmt.Sprintf(`
	%s

	resource "%s" %s {
		database_name         = "%s"
		table_name            = "${adx_table.test.name}"
		enabled 		      = true
		hint_allocated_rate   = 2.1
	}
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.DatabaseName)
}

func (this ADXTableStreamingIngestionPolicyTestResource) template(rtc *ResourceTestContext[TableStreamingIngestionPolicy]) string {
	return fmt.Sprintf(`
	resource "adx_table" "test" {
		database_name = "%s"
		name          = "%s"
		table_schema  = "f1:string,f2:string,f4:string,f3:int"
	}
	`, rtc.DatabaseName, rtc.EntityName)
}
