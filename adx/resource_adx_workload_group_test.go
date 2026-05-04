package adx

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ADXWorkloadGroupTestResource struct{}

func TestAccADXWorkloadGroup_basic(t *testing.T) {
	var entity ADXWorkloadGroup
	r := ADXWorkloadGroupTestResource{}
	rtcBuilder := BuildResourceTestContext[ADXWorkloadGroup]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_workload_group").
		DatabaseName(testAccDatabaseName()).
		EntityType("workload_group").
		ReadStatementFunc(func(id string) (string, error) {
			wgId, err := parseADXWorkloadGroupID(id)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(".show workload_group %s", escapeEntityNameIfRequired(wgId.Name)), nil
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
				),
			},
			{
				Config: r.update(rtc),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
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

func TestAccADXWorkloadGroup_full(t *testing.T) {
	var entity ADXWorkloadGroup
	r := ADXWorkloadGroupTestResource{}
	rtcBuilder := BuildResourceTestContext[ADXWorkloadGroup]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_workload_group").
		DatabaseName(testAccDatabaseName()).
		EntityType("workload_group").
		ReadStatementFunc(func(id string) (string, error) {
			wgId, err := parseADXWorkloadGroupID(id)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(".show workload_group %s", escapeEntityNameIfRequired(wgId.Name)), nil
		}).Build()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				Config: r.full(rtc),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
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

func (this ADXWorkloadGroupTestResource) basic(rtc *ResourceTestContext[ADXWorkloadGroup]) string {
	return fmt.Sprintf(`
	resource "%s" %s {
		database_name = "%s"
		name          = "%s"

		request_rate_limit_policies = jsonencode([
			{
				IsEnabled = true
				Scope     = "WorkloadGroup"
				LimitKind = "ConcurrentRequests"
				Properties = {
					MaxConcurrentRequests = 100
				}
			}
		])
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName)
}

func (this ADXWorkloadGroupTestResource) update(rtc *ResourceTestContext[ADXWorkloadGroup]) string {
	return fmt.Sprintf(`
	resource "%s" %s {
		database_name = "%s"
		name          = "%s"

		request_rate_limit_policies = jsonencode([
			{
				IsEnabled = true
				Scope     = "WorkloadGroup"
				LimitKind = "ConcurrentRequests"
				Properties = {
					MaxConcurrentRequests = 50
				}
			},
			{
				IsEnabled = true
				Scope     = "Principal"
				LimitKind = "ConcurrentRequests"
				Properties = {
					MaxConcurrentRequests = 10
				}
			}
		])
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName)
}

func (this ADXWorkloadGroupTestResource) full(rtc *ResourceTestContext[ADXWorkloadGroup]) string {
	return fmt.Sprintf(`
	resource "%s" %s {
		database_name = "%s"
		name          = "%s"

		request_limits_policy = jsonencode({
			DataScope = {
				IsRelaxable = true
				Value       = "HotCache"
			}
			MaxMemoryPerQueryPerNode = {
				IsRelaxable = true
				Value       = 6442450944
			}
			MaxMemoryPerIterator = {
				IsRelaxable = true
				Value       = 5368709120
			}
			MaxFanoutThreadsPercentage = {
				IsRelaxable = true
				Value       = 100
			}
			MaxFanoutNodesPercentage = {
				IsRelaxable = true
				Value       = 100
			}
			MaxResultRecords = {
				IsRelaxable = true
				Value       = 500000
			}
			MaxResultBytes = {
				IsRelaxable = true
				Value       = 67108864
			}
			MaxExecutionTime = {
				IsRelaxable = true
				Value       = "00:04:00"
			}
		})

		request_rate_limit_policies = jsonencode([
			{
				IsEnabled = true
				Scope     = "WorkloadGroup"
				LimitKind = "ConcurrentRequests"
				Properties = {
					MaxConcurrentRequests = 100
				}
			},
			{
				IsEnabled = true
				Scope     = "Principal"
				LimitKind = "ConcurrentRequests"
				Properties = {
					MaxConcurrentRequests = 25
				}
			}
		])

		request_rate_limits_enforcement_policy = jsonencode({
			QueriesEnforcementLevel  = "QueryHead"
			CommandsEnforcementLevel = "Database"
		})

		request_queuing_policy = jsonencode({
			IsEnabled = true
		})

		query_consistency_policy = jsonencode({
			QueryConsistency = {
				IsRelaxable = true
				Value       = "Weak"
			}
			CachedResultsMaxAge = {
				IsRelaxable = true
				Value       = "00:05:00"
			}
		})
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName)
}
