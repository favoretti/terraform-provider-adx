package adx

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// A soft delete period of 0 is explicitly supported by ADX in table and materialized view
// level retention policies (but not at database level). Ingested data is then never committed
// to the entity, which is why ADX also requires recoverability to be disabled, and requires
// every update policy referencing the entity as its source to be transactional.
// See https://learn.microsoft.com/en-us/kusto/management/retention-policy
var softDeletePeriodRegex = regexp.MustCompile(`^(0|[1-9]\d{0,4}|[1-2]\d{5}|3[0-4]\d{4}|35\d{4}|36[0-3]\d{3}|364[0-5]\d{2}|3646[0-2]\d|36463[0-5])[dhms]$`)

const softDeletePeriodValidationMessage = "soft delete timespan must be in the format of <amount><unit> such as 1m (one minute) or 30d (thirty days), maximum is 364635d (999 years). Use 0 (such as 0s) to keep no data at all"

func validateSoftDeletePeriod() schema.SchemaValidateDiagFunc {
	return validate.StringMatch(softDeletePeriodRegex, softDeletePeriodValidationMessage)
}

// isZeroSoftDeletePeriod reports whether the given soft delete period means "keep no data".
// The regex only accepts a bare zero (0s, 0d, ...), so a leading zero is unambiguous.
func isZeroSoftDeletePeriod(period string) bool {
	return softDeletePeriodRegex.MatchString(period) && strings.HasPrefix(period, "0")
}

// validateRetentionPolicyRecoverability enforces the ADX constraint that recoverability can
// only be disabled when the soft delete period is 0. The transactional requirement on update
// policies referencing the entity is deliberately not checked here: those policies live on
// other entities, may use a wildcard source and may not be managed by Terraform at all, so
// that error is left to ADX.
func validateRetentionPolicyRecoverability(softDeletePeriod string, recoverability bool) error {
	if isZeroSoftDeletePeriod(softDeletePeriod) && recoverability {
		return fmt.Errorf("recoverability must be false when soft_delete_period is %q: ADX does not commit ingested data when the soft delete period is 0, so recoverability can only be disabled", softDeletePeriod)
	}
	return nil
}

func retentionPolicyCustomDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if err := clusterConfigCustomDiff(ctx, diff, meta); err != nil {
		return err
	}

	// Values that are still unknown at plan time come back as their zero value and are skipped.
	return validateRetentionPolicyRecoverability(diff.Get("soft_delete_period").(string), diff.Get("recoverability").(bool))
}
