package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXTableRestrictedViewPolicyTestResource struct{}

func TestAccADXTableRestrictedViewPolicy_basic(t *testing.T) {
	var entity TableRestrictedViewPolicy
	r := ADXTableRestrictedViewPolicyTestResource{}
	rtcBuilder := BuildResourceTestContext[TableRestrictedViewPolicy]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_table_RestrictedView_policy").
		DatabaseName("test-db").
		EntityType("RestrictedView").
		ReadStatementFunc(GetAccTestPolicyReadStatementFunc()).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				Config: r.basic(rtc, "true"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "enabled", "true"),
				),
			},
			{
				Config: r.basic(rtc, "false"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "enabled", "false"),
				),
			},
			{
				ResourceName:            rtc.GetTFName(),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled", "follower_database"},
			},
		},
	})
}

// Requires a follower database already configured with table `sample_shared_table`
func TestAccADXTableRestrictedViewPolicy_follower(t *testing.T) {
	var entity TableRestrictedViewPolicy
	r := ADXTableRestrictedViewPolicyTestResource{}
	rtcBuilder := BuildResourceTestContext[TableRestrictedViewPolicy]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_table_RestrictedView_policy").
		DatabaseName("shareable-db").
		EntityType("RestrictedView").
		EntityName("sample_shared_table").
		ReadStatementFunc(GetAccTestPolicyReadStatementFunc()).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: r.follower(rtc, rtc.EntityName, "true"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "enabled", "true"),
				),
			},
			{
				Config: r.follower(rtc, rtc.EntityName, "false"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "enabled", "false"),
				),
			},
			{
				ResourceName:            rtc.GetTFName(),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled", "follower_database"},
			},
		},
	})
}

func (this ADXTableRestrictedViewPolicyTestResource) basic(rtc *ResourceTestContext[TableRestrictedViewPolicy], enabled string) string {
	return fmt.Sprintf(`
	%s

	resource "%s" %s {
		database_name = "%s"
		table_name    = "${adx_table.test.name}"
		enabled = "%s"
	}
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.DatabaseName, enabled)
}

func (this ADXTableRestrictedViewPolicyTestResource) follower(rtc *ResourceTestContext[TableRestrictedViewPolicy], tableName string, enabled string) string {
	return fmt.Sprintf(`

	resource "%s" %s {
		database_name     = "%s"
		table_name        = "%s"
		enabled     = "%s"
		follower_database = true
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, tableName, enabled)
}

func (this ADXTableRestrictedViewPolicyTestResource) template(rtc *ResourceTestContext[TableRestrictedViewPolicy]) string {
	return fmt.Sprintf(`
	resource "adx_table" "test" {
		database_name = "%s"
		name          = "%s"
		table_schema  = "f1:string,f2:string,f4:string,f3:int"
	}
	`, rtc.DatabaseName, rtc.EntityName)
}
