package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXMaterializedViewCachingPolicyTestResource struct{}

func TestAccADXMaterializedViewCachingPolicy_basic(t *testing.T) {
	var entity MaterializedViewCachingPolicy
	r := ADXMaterializedViewCachingPolicyTestResource{}
	rtcBuilder := BuildResourceTestContext[MaterializedViewCachingPolicy]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_materialized_view_caching_policy").
		DatabaseName("test-db").
		EntityType("caching").
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
					resource.TestCheckResourceAttr(rtc.GetTFName(), "view_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "data_hot_span", "3d"),
				),
			},
		},
	})
}

// Requires a follower database already configured with view `sample_shared_mv`
func TestAccADXMaterializedViewCachingPolicy_follower(t *testing.T) {
	var entity MaterializedViewCachingPolicy
	r := ADXMaterializedViewCachingPolicyTestResource{}
	rtcBuilder := BuildResourceTestContext[MaterializedViewCachingPolicy]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_materialized_view_caching_policy").
		DatabaseName("shareable-db").
		EntityType("caching").
		EntityName("sample_shared_mv").
		ReadStatementFunc(GetAccTestPolicyReadStatementFunc()).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: r.follower(rtc, rtc.EntityName),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "view_name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "data_hot_span", "3d"),
				),
			},
		},
	})
}

func (this ADXMaterializedViewCachingPolicyTestResource) basic(rtc *ResourceTestContext[MaterializedViewCachingPolicy]) string {
	return fmt.Sprintf(`
	%s

	resource "%s" %s {
		database_name = "%s"
		view_name    = "${adx_materialized_view.test.name}"
		data_hot_span = "3d"
	}
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.DatabaseName)
}

func (this ADXMaterializedViewCachingPolicyTestResource) follower(rtc *ResourceTestContext[MaterializedViewCachingPolicy], viewName string) string {
	return fmt.Sprintf(`

	resource "%s" %s {
		database_name     = "%s"
		view_name         = "%s"
		data_hot_span     = "3d"
		follower_database = true
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, viewName)
}

func (this ADXMaterializedViewCachingPolicyTestResource) template(rtc *ResourceTestContext[MaterializedViewCachingPolicy]) string {
	return fmt.Sprintf(`
	resource "adx_materialized_view" "test" {
		database_name     = "%s"
		name              = "%s"
		query             = "${adx_table.test.name} | summarize arg_max(f3, *) by f1"
		source_table_name = adx_table.test.name
	}

	resource "adx_table" "test" {
		database_name = "%s"
		name          = "table_for_mv_%s"
		table_schema  = "f1:string,f2:string,f4:string,f3:int"
	}
	`, rtc.DatabaseName, rtc.EntityName, rtc.DatabaseName, acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
}
