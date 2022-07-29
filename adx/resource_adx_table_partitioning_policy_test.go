package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXTablePartitioningPolicyTestResource struct{}

func TestAccADXTablePartitioningPolicy_basic(t *testing.T) {
	var entity TablePartitioningPolicy
	r := ADXTablePartitioningPolicyTestResource{}
	rtcBuilder := BuildResourceTestContext[TablePartitioningPolicy]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_table_partitioning_policy").
		DatabaseName("test-db").
		EntityType("partitioning").
		ReadStatementFunc(GetAccTestPolicyReadStatementFunc("table","partitioning")).
		IDParserFunc(GetAccTestPolicyIDParserFunc()).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				Config: r.basic(rtc,64),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					//resource.TestCheckResourceAttrSet(rtc.GetTFName(), "partition_key"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "partition_key.0.column_name", "f1"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "partition_key.0.kind", "Hash"),
					//resource.TestCheckResourceAttrSet(rtc.GetTFName(), "partition_key.0.hash_properties"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "partition_key.0.hash_properties.0.function", "XxHash64"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "partition_key.0.hash_properties.0.max_partition_count", "64"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "partition_key.0.hash_properties.0.seed", "2"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "partition_key.0.hash_properties.0.partition_assignment_mode", "Uniform"),
				),
			},
			{
				Config: r.basic(rtc,128),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					//resource.TestCheckResourceAttrSet(rtc.GetTFName(), "partition_key"),
					//resource.TestCheckResourceAttrSet(rtc.GetTFName(), "partition_key.0.hash_properties"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "partition_key.0.hash_properties.0.max_partition_count", "128"),
				),
			},
			{
				Config: r.uniformrange(rtc),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					//resource.TestCheckResourceAttrSet(rtc.GetTFName(), "partition_key"),
					//resource.TestCheckResourceAttrSet(rtc.GetTFName(), "partition_key.0.uniform_range_properties"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "partition_key.0.uniform_range_properties.0.range_size", "2.00:00:00"),
				),
			},
		},
	})
}

func (this ADXTablePartitioningPolicyTestResource) basic(rtc *ResourceTestContext[TablePartitioningPolicy], partCount int) string {
	return fmt.Sprintf(`
	%s

	resource "%s" %s {
		database_name         = "%s"
		table_name            = "${adx_table.test.name}"
		effective_date_time = "2022-07-19T13:56:45Z"

		partition_key {
			column_name = "f1"
			kind        = "Hash"

			hash_properties {
				function                  = "XxHash64"
				max_partition_count       = %d
				seed                      = 2
				partition_assignment_mode = "Uniform"
			}
		}
	}
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.DatabaseName, partCount)
}

func (this ADXTablePartitioningPolicyTestResource) uniformrange(rtc *ResourceTestContext[TablePartitioningPolicy]) string {
	return fmt.Sprintf(`
	%s

	resource "%s" %s {
		database_name         = "%s"
		table_name            = "${adx_table.test.name}"
		effective_date_time = "2022-07-19T13:56:45Z"

		partition_key {
			column_name = "f3"
			kind        = "UniformRange"

			uniform_range_properties {
				range_size                = "2.00:00:00"
				reference                 = "1990-01-01T00:00:00"
				override_creation_time    = true
			}
		}
	}
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.DatabaseName)
}

func (this ADXTablePartitioningPolicyTestResource) template(rtc *ResourceTestContext[TablePartitioningPolicy]) string {
	return fmt.Sprintf(`
	resource "adx_table" "test" {
		database_name = "%s"
		name          = "%s"
		table_schema  = "f1:string,f2:string,f4:string,f3:datetime"
	}
	`, rtc.DatabaseName, rtc.EntityName)
}
