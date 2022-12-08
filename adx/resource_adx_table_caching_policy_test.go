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
		ReadStatementFunc(GetAccTestPolicyReadStatementFunc()).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				Config: r.basic(rtc, "3d"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "data_hot_span", "3d"),
				),
			},
			{
				Config: r.basic(rtc, "1d"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "data_hot_span", "1d"),
				),
			},
			{
				ResourceName:            rtc.GetTFName(),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"data_hot_span", "follower_database"},
			},
		},
	})
}

// Requires a follower database already configured with table `sample_shared_table`
func TestAccADXTableCachingPolicy_follower(t *testing.T) {
	var entity TableCachingPolicy
	r := ADXTableCachingPolicyTestResource{}
	rtcBuilder := BuildResourceTestContext[TableCachingPolicy]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_table_caching_policy").
		DatabaseName("shareable-db").
		EntityType("caching").
		EntityName("sample_shared_table").
		ReadStatementFunc(GetAccTestPolicyReadStatementFunc()).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: r.follower(rtc, rtc.EntityName, "3d"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "data_hot_span", "3d"),
				),
			},
			{
				Config: r.follower(rtc, rtc.EntityName, "1d"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "data_hot_span", "1d"),
				),
			},
			{
				ResourceName:            rtc.GetTFName(),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"data_hot_span", "follower_database"},
			},
		},
	})
}

func (this ADXTableCachingPolicyTestResource) basic(rtc *ResourceTestContext[TableCachingPolicy], hotCache string) string {
	return fmt.Sprintf(`
	%s

	resource "%s" %s {
		database_name = "%s"
		table_name    = "${adx_table.test.name}"
		data_hot_span = "%s"
	}
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.DatabaseName, hotCache)
}

func (this ADXTableCachingPolicyTestResource) follower(rtc *ResourceTestContext[TableCachingPolicy], tableName string, hotCache string) string {
	return fmt.Sprintf(`

	resource "%s" %s {
		database_name     = "%s"
		table_name        = "%s"
		data_hot_span     = "%s"
		follower_database = true
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, tableName, hotCache)
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
