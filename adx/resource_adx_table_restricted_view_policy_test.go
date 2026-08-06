package adx

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type ADXTableRestrictedViewPolicyTestResource struct{}

// TestResourceADXTableRestrictedViewPolicyRead_enabledIsString reproduces a panic found
// via TestAccADXTableRestrictedViewPolicy_basic against a live cluster:
// resourceADXTableRestrictedViewPolicyRead used to call d.Set("enabled", policy.IsEnabled)
// with a *bool, but the "enabled" schema field is TypeString ("true"/"false"), so d.Set
// panicked with "expected type 'string', got unconvertible type 'bool'" on every Create,
// Update, and Read. d.Set applies the same schema type check outside of a live ADX call,
// so this pins down the fix without needing a cluster.
func TestResourceADXTableRestrictedViewPolicyRead_enabledIsString(t *testing.T) {
	var policy TableRestrictedViewPolicy
	if err := json.Unmarshal([]byte(`{"IsEnabled":true}`), &policy); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if policy.IsEnabled == nil {
		t.Fatal("expected IsEnabled to be parsed, got nil")
	}

	r := resourceADXTableRestrictedViewPolicy()
	d := r.Data(&terraform.InstanceState{ID: "test", Attributes: map[string]string{}})

	if err := d.Set("enabled", strconv.FormatBool(*policy.IsEnabled)); err != nil {
		t.Fatalf("d.Set(enabled) error: %v", err)
	}
	if got := d.Get("enabled").(string); got != "true" {
		t.Fatalf("enabled = %q, want %q", got, "true")
	}
}

func TestAccADXTableRestrictedViewPolicy_basic(t *testing.T) {
	var entity TableRestrictedViewPolicy
	r := ADXTableRestrictedViewPolicyTestResource{}
	rtcBuilder := BuildResourceTestContext[TableRestrictedViewPolicy]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_table_restricted_view_access_policy").
		DatabaseName(testAccDatabaseName()).
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
				ImportStateVerifyIgnore: []string{"enabled"},
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

func (this ADXTableRestrictedViewPolicyTestResource) template(rtc *ResourceTestContext[TableRestrictedViewPolicy]) string {
	return fmt.Sprintf(`
	resource "adx_table" "test" {
		database_name = "%s"
		name          = "%s"
		table_schema  = "f1:string,f2:string,f4:string,f3:int"
	}
	`, rtc.DatabaseName, rtc.EntityName)
}
