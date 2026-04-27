package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXMergePolicyTestResource struct{}

func TestAccADXMergePolicy_table(t *testing.T) {
	var entity TablePolicy
	r := ADXMergePolicyTestResource{}
	rtcBuilder := BuildResourceTestContext[TablePolicy]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_merge_policy").
		DatabaseName(testAccDatabaseName()).
		EntityType("merge").
		ReadStatementFunc(GetAccTestPolicyReadStatementFunc()).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				Config: r.tableBasic(rtc, 16000000, 30000, 100, 24, true, true),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "entity_type", "table"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "row_count_upper_bound_for_merge", "16000000"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "original_size_mb_upper_bound_for_merge", "30000"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "max_extents_to_merge", "100"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "max_range_in_hours", "24"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "allow_rebuild", "true"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "allow_merge", "true"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "lookback_kind", "Default"),
				),
			},
			{
				Config: r.tableBasic(rtc, 16000000, 30000, 100, 48, false, true),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "max_range_in_hours", "48"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "allow_rebuild", "false"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "allow_merge", "true"),
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

func (this ADXMergePolicyTestResource) tableBasic(rtc *ResourceTestContext[TablePolicy], rowCount int, originalSize int, maxExtents int, maxRange int, allowRebuild bool, allowMerge bool) string {
	return fmt.Sprintf(`
	%s

	resource "%s" %s {
		database_name                          = "%s"
		entity_type                            = "table"
		entity_name                            = "${adx_table.test.name}"
		row_count_upper_bound_for_merge        = %d
		original_size_mb_upper_bound_for_merge = %d
		max_extents_to_merge                   = %d
		max_range_in_hours                     = %d
		allow_rebuild                          = %t
		allow_merge                            = %t
		lookback_kind                          = "Default"
	}
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.DatabaseName, rowCount, originalSize, maxExtents, maxRange, allowRebuild, allowMerge)
}

func (this ADXMergePolicyTestResource) template(rtc *ResourceTestContext[TablePolicy]) string {
	return fmt.Sprintf(`
	resource "adx_table" "test" {
		database_name = "%s"
		name          = "%s"
		table_schema  = "f1:string,f2:string,f4:string,f3:int"
	}
	`, rtc.DatabaseName, rtc.EntityName)
}
