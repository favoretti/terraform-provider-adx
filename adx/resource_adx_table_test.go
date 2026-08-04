package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

// checkTableSchemaMatches compares table_schema after unescaping both sides (e.g.
// "['version']" vs "version"). Whether ADX's `.show table ... cslschema` escapes a given
// column name is a property of the Kusto engine version running on the cluster, not of
// this provider, so asserting on the raw escaped/unescaped form directly is brittle
// across clusters -- this mirrors the table_schema DiffSuppressFunc's own comparison.
func checkTableSchemaMatches(tfName, expectedUnescaped string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[tfName]
		if !ok {
			return fmt.Errorf("not found: %s", tfName)
		}
		got := unescapeTableSchema(rs.Primary.Attributes["table_schema"])
		if got != expectedUnescaped {
			return fmt.Errorf("table_schema (unescaped) = %q, want %q (raw state value: %q)", got, expectedUnescaped, rs.Primary.Attributes["table_schema"])
		}
		return nil
	}
}

type ADXTableTestResource struct{}

func TestAccADXTable_basic(t *testing.T) {
	var entity TableSchema
	r := ADXTableTestResource{}
	rtcBuilder := BuildResourceTestContext[TableSchema]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_table").
		DatabaseName(testAccDatabaseName()).
		EntityType("table").
		ReadStatementFunc(func(id string) (string, error) {
			funcId, err := parseADXTableID(id)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(".show tables | where TableName == '%s'", funcId.Name), nil
		}).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				Config: r.basic(rtc),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "column.0.name", "f1"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_schema", "f1:string,f2:string,f3:int"),
				),
			},
			{
				Config: r.basic_update(rtc),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "column.0.name", "f1"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "column.3.name", "f4"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "table_schema", "f1:string,f2:string,f3:int,f4:string"),
				),
			},
		},
	})
}

func TestAccADXTable_basic_inline(t *testing.T) {
	var entity TableSchema
	r := ADXTableTestResource{}
	rtcBuilder := BuildResourceTestContext[TableSchema]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_table").
		DatabaseName(testAccDatabaseName()).
		EntityType("table").
		ReadStatementFunc(func(id string) (string, error) {
			funcId, err := parseADXTableID(id)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(".show tables | where TableName == '%s'", funcId.Name), nil
		}).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				Config: r.basic_inline(rtc),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					checkTableSchemaMatches(rtc.GetTFName(), "f1:string,version:string,f4:string"),
				),
			},
			{
				Config: r.basic_inline_update(rtc),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					checkTableSchemaMatches(rtc.GetTFName(), "f1:string,version:string,f4:string,f3:int"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "docstring", "This is table"),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "folder", "iamafolder"),
				),
			},
		},
	})
}

func (this ADXTableTestResource) basic_inline(rtc *ResourceTestContext[TableSchema]) string {
	return fmt.Sprintf(`

	resource "%s" %s {
		database_name = "%s"
		name          = "%s"
		table_schema  = "f1:string,version:string,f4:string"
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName)
}

func (this ADXTableTestResource) basic_inline_update(rtc *ResourceTestContext[TableSchema]) string {
	return fmt.Sprintf(`

	resource "%s" %s {
		database_name = "%s"
		name          = "%s"
		table_schema  = "f1:string,version:string,f4:string,f3:int"
		docstring     = "This is table"
		folder        = "iamafolder"
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName)
}

func (this ADXTableTestResource) basic(rtc *ResourceTestContext[TableSchema]) string {
	return fmt.Sprintf(`

	resource "%s" %s {
		database_name = "%s"
		name          = "%s"
		
		column {
			name = "f1"
			type = "string"
		}
	
		column {
			name = "f2"
			type = "string"
		}
	
		column {
			name = "f3"
			type = "int"
		}
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName)
}

func (this ADXTableTestResource) basic_update(rtc *ResourceTestContext[TableSchema]) string {
	return fmt.Sprintf(`

	resource "%s" %s {
		database_name = "%s"
		name          = "%s"
		
		column {
			name = "['f1']"
			type = "string"
		}
	
		column {
			name = "f2"
			type = "string"
		}
	
		column {
			name = "f3"
			type = "int"
		}

		column {
			name = "f4"
			type = "string"
		}
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName)
}

func TestADXTable_unescapeTableSchema(t *testing.T) {
	name := unescapeTableSchema("name:string, ename:int,othername:string")
	assert.Equal(t, "name:string, ename:int,othername:string", name, "an already-unescaped name should have been left alone")

	name = unescapeTableSchema("name:string, ['ename']:int,othername:string")
	assert.Equal(t, "name:string, ename:int,othername:string", name, "ename should have had adx escaping removed")
}
