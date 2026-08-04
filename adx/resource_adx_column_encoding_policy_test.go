package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXColumnEncodingPolicyTestResource struct{}

func TestAccADXColumnEncodingPolicy_basic(t *testing.T) {
	var entity ColumnEncodingPolicy
	r := ADXColumnEncodingPolicyTestResource{}
	rtcBuilder := BuildResourceTestContext[ColumnEncodingPolicy]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_column_encoding_policy").
		DatabaseName(testAccDatabaseName()).
		EntityType("encoding").
		ReadStatementFunc(GetAccTestPolicyReadStatementFunc()).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: r.basic(rtc, fmt.Sprintf("%s.f1", rtc.EntityName), "BigObject"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "entity_identifier", fmt.Sprintf("%s.f1", rtc.EntityName)),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "encoding_policy_type", "BigObject"),
				),
			},
			{
				// Re-apply the same values: this step's Check asserts the same
				// entity_identifier/encoding_policy_type as step 1, so Config must match
				// (previously passed mismatched values copy-pasted from a different
				// test, which this step's own checks would have failed once step 1's
				// missing depends_on stopped masking it).
				Config: r.basic(rtc, fmt.Sprintf("%s.f1", rtc.EntityName), "BigObject"),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "entity_identifier", fmt.Sprintf("%s.f1", rtc.EntityName)),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "encoding_policy_type", "BigObject"),
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

func (this ADXColumnEncodingPolicyTestResource) basic(rtc *ResourceTestContext[ColumnEncodingPolicy], entityIdentifier string, encodingPolicyType string) string {
	return fmt.Sprintf(`
	%s

	resource "%s" %s {
		database_name         	= "%s"
		entity_identifier 		= "%s"
		encoding_policy_type   	= "%s"

		depends_on = [adx_table.test]
	}
	`, this.template(rtc), rtc.Type, rtc.Label, rtc.DatabaseName, entityIdentifier, encodingPolicyType)
}

func (this ADXColumnEncodingPolicyTestResource) template(rtc *ResourceTestContext[ColumnEncodingPolicy]) string {
	return fmt.Sprintf(`
	resource "adx_table" "test" {
		database_name = "%s"
		name          = "%s"
		table_schema  = "f1:dynamic,f2:string,f4:string,f3:int"
	}
	`, rtc.DatabaseName, rtc.EntityName)
}
