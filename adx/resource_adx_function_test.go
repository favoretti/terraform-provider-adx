package adx

import (
	"os"
	"testing"
)

/*func TestAccFunction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
	  PreCheck:     func() { testAccPreCheck(t) },
	  Providers:    testAccProviders,
	  CheckDestroy: testAccCheckExampleResourceDestroy,
	  Steps: []resource.TestStep{
		{
		  Config: testAccExampleResource(rName),
		  Check: resource.ComposeTestCheckFunc(
			testAccCheckExampleResourceExists("example_widget.foo", &widgetBefore),
		  ),
		},
		{
		  Config: testAccExampleResource_removedPolicy(rName),
		  Check: resource.ComposeTestCheckFunc(
			testAccCheckExampleResourceExists("example_widget.foo", &widgetAfter),
		  ),
		},
	  },
	})
  }*/