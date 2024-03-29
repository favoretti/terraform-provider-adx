package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXTableIngestionBatchingPolicyTestResource struct{}

func TestAccADXTableIngestionBatchingPolicy_basic(t *testing.T) {
	var entity TableIngestionBatchingPolicy
	r := ADXTableIngestionBatchingPolicyTestResource{}
	rtcBuilder := BuildResourceTestContext[TableIngestionBatchingPolicy]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_table_ingestion_batching_policy").
		DatabaseName("test-db").
		EntityType("ingestionbatching").
		ReadStatementFunc(GetAccTestPolicyReadStatementFunc()).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				Config: r.basic(rtc, "00:10:00", "100"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "max_batching_timespan", "00:10:00"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "max_items", "30000"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "max_raw_size_mb", "100"),
				),
			},
			{
				Config: r.basic(rtc, "00:05:00", "200"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "max_batching_timespan", "00:05:00"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "max_items", "30000"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "max_raw_size_mb", "200"),
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

func (this ADXTableIngestionBatchingPolicyTestResource) basic(rtc *ResourceTestContext[TableIngestionBatchingPolicy], timespan string, maxSize string) string {
	return fmt.Sprintf(`
	%s

	resource "%s" %s {
		database_name         = "%s"
		table_name            = "${adx_table.test.name}"
		max_batching_timespan = "%s"
		max_items             = 30000
		max_raw_size_mb       = %s
	}
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.DatabaseName, timespan, maxSize)
}

func (this ADXTableIngestionBatchingPolicyTestResource) template(rtc *ResourceTestContext[TableIngestionBatchingPolicy]) string {
	return fmt.Sprintf(`
	resource "adx_table" "test" {
		database_name = "%s"
		name          = "%s"
		table_schema  = "f1:string,f2:string,f4:string,f3:int"
	}
	`, rtc.DatabaseName, rtc.EntityName)
}
