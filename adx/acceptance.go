package adx

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-kusto-go/kusto"
	
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func GetTestClusterConfig() *ClusterConfig {
	return &ClusterConfig{
		ClientID:     os.Getenv("ADX_CLIENT_ID"),
		ClientSecret: os.Getenv("ADX_CLIENT_SECRET"),
		TenantID:     os.Getenv("ADX_TENANT_ID"),
		URI:          os.Getenv("ADX_ENDPOINT"),
	}
}

type ResourceTestContext[T any] struct {
	Test          *testing.T
	Cluster       *ClusterConfig
	DatabaseName  string
	EntityType    string
	EntityName    string
	Type          string
	Label         string
	ReadStatement string
}

type ResourceTestContextBuilder[T any] struct {
	context 			  *ResourceTestContext[T]
	interpolateEntityName bool
}

func (this *ResourceTestContextBuilder[T]) Build() (*ResourceTestContext[T], error) {
	if this.context.Cluster==nil {
		this.context.Cluster = GetTestClusterConfig()
	}
	if this.context.Label=="" {
		this.context.Label = "test"
	}
	if this.context.Test==nil {
		return nil, fmt.Errorf("Test cannot be nil")
	}
	if this.context.DatabaseName=="" {
		return nil, fmt.Errorf("DatabaseName cannot be empty")
	}
	if this.context.EntityType=="" {
		return nil, fmt.Errorf("EntityType cannot be empty")
	}
	if this.context.EntityName=="" {
		this.context.EntityName = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	}
	if this.context.Type=="" {
		return nil, fmt.Errorf("Type cannot be empty")
	}
	if this.context.ReadStatement=="" {
		return nil, fmt.Errorf("ReadStatement cannot be empty")
	} else if this.interpolateEntityName {
		this.context.ReadStatement = fmt.Sprintf(this.context.ReadStatement,this.context.EntityName)
	}
	return this.context, nil
}

func (this *ResourceTestContextBuilder[T]) Initialize() *ResourceTestContextBuilder[T] {
	this.context = &ResourceTestContext[T]{}
	return this
}

func (this *ResourceTestContextBuilder[T]) Test(test  *testing.T) *ResourceTestContextBuilder[T] {
	this.context.Test = test
	return this
}

func (this *ResourceTestContextBuilder[T]) Cluster(cluster *ClusterConfig) *ResourceTestContextBuilder[T] {
	this.context.Cluster = cluster
	return this
}

func (this *ResourceTestContextBuilder[T]) DatabaseName(value string) *ResourceTestContextBuilder[T] {
	this.context.DatabaseName = value
	return this
}

func (this *ResourceTestContextBuilder[T]) EntityType(value string) *ResourceTestContextBuilder[T] {
	this.context.EntityType = value
	return this
}

func (this *ResourceTestContextBuilder[T]) EntityName(value string) *ResourceTestContextBuilder[T] {
	this.context.EntityName = value
	return this
}

func (this *ResourceTestContextBuilder[T]) Type(value string) *ResourceTestContextBuilder[T] {
	this.context.Type = value
	return this
}

func (this *ResourceTestContextBuilder[T]) Label(value string) *ResourceTestContextBuilder[T] {
	this.context.Label = value
	return this
}

func (this *ResourceTestContextBuilder[T]) ReadStatement(value string, interpolateEntityName bool) *ResourceTestContextBuilder[T] {
	this.context.ReadStatement = value
	this.interpolateEntityName = interpolateEntityName
	return this
}

func BuildResourceTestContext[T any]() *ResourceTestContextBuilder[T] {
	builder := &ResourceTestContextBuilder[T]{}
	return builder.Initialize()
}

func (this *ResourceTestContext[T]) GetTFName() string {
	return fmt.Sprintf("%s.%s", this.Type, this.Label)
}

func (this *ResourceTestContext[T]) GetADXClient() (*kusto.Client, error) {
	return getADXClient(testAccProvider.Meta(), this.Cluster)
}

func (this *ResourceTestContext[T]) GetADXEntity() (*T, error) {
	entities, err := queryADXMgmtAndParse[T](context.Background(), testAccProvider.Meta(), this.Cluster, this.DatabaseName, this.ReadStatement)
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 {
		return nil, nil
	} else if len(entities) > 1 {
		return nil, fmt.Errorf("ADX returned too many rows for entity read query (%s) (%s)",this.EntityName, this.GetTFName())
	}
	return &entities[0], nil
}

func (this *ResourceTestContext[T]) GetTestCheckEntityExists(entity *T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[this.GetTFName()]
		if !ok {
			return fmt.Errorf("Not found: %s", this.GetTFName())
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("ID is not set for (%s) (%s)",this.EntityName, this.GetTFName())
		}
		result, err := this.GetADXEntity()
		if err != nil {
			return fmt.Errorf("Failed to retrieve entity from ADX (%s) (%s): %+v", this.EntityName, this.GetTFName(), err)
		}
		*entity = *result
		return nil
	}
}

func (this *ResourceTestContext[T]) GetTestCheckEntityDestroyed() func(*terraform.State) error {
	return func(s *terraform.State) error {
		// loop through the resources in state, verifying each entity is destroyed
		for _, rs := range s.RootModule().Resources {
			// Ignore resources which are not of the correct type
			if rs.Type != this.Type {
				continue
			}
			adxResourceId, _ := parseADXFunctionID(rs.Primary.ID)
			err := this.CheckEntityDestroyed(adxResourceId.Name)
			if err != nil {
				return fmt.Errorf("%+v. ID: %s", err, rs.Primary.ID)
			}
		}
		return nil
	}
}

func (this *ResourceTestContext[T]) CheckEntityDestroyed(entityNameOverride string) error {
	entityName := this.EntityName
	if entityNameOverride != "" {
		entityName = entityNameOverride
	}

	entity, err := this.GetADXEntity()
	if err != nil {
		return fmt.Errorf("Failed to check entity destroyed in ADX (%s) (%s): %+v", entityName, this.GetTFName(), err)
	}
	if entity != nil {
		return fmt.Errorf("Entity (%s) of type (%s) not destroyed in ADX", entityName, this.EntityType)
	}
	return nil
}

func testAccPreCheck(t *testing.T) {
	if err := os.Getenv("ADX_CLIENT_ID"); err == "" {
		t.Fatal("ADX_CLIENT_ID must be set for acceptance tests")
	}

	if err := os.Getenv("ADX_CLIENT_SECRET"); err == "" {
		t.Fatal("ADX_CLIENT_SECRET must be set for acceptance tests")
	}

	if err := os.Getenv("ADX_TENANT_ID"); err == "" {
		t.Fatal("ADX_TENANT_ID must be set for acceptance tests")
	}

	if err := os.Getenv("ADX_ENDPOINT"); err == "" {
		t.Fatal("ADX_ENDPOINT must be set for acceptance tests")
	}
}