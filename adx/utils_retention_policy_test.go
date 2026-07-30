package adx

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestValidateSoftDeletePeriod(t *testing.T) {
	cases := []struct {
		input string
		valid bool
	}{
		// A soft delete period of 0 keeps no data at all and is explicitly supported by ADX
		// for table and materialized view retention policies.
		{"0s", true},
		{"0m", true},
		{"0h", true},
		{"0d", true},
		{"1m", true},
		{"500m", true},
		{"30d", true},
		{"364635d", true},

		{"00s", false},
		{"01s", false},
		{"0", false},
		{"0sec", false},
		{"00:00:00", false},
		{"-1s", false},
		{"0x", false},
		{"364636d", false},
		{"", false},
	}

	validator := validateSoftDeletePeriod()

	for _, tc := range cases {
		diags := validator(tc.input, cty.Path{})
		if valid := !diags.HasError(); valid != tc.valid {
			t.Errorf("soft_delete_period %q: expected valid=%t, got valid=%t (%v)", tc.input, tc.valid, valid, diags)
		}
	}
}

func TestValidateRetentionPolicyRecoverability(t *testing.T) {
	cases := []struct {
		softDeletePeriod string
		recoverability   bool
		expectError      bool
	}{
		// ADX does not commit ingested data when the soft delete period is 0, so
		// recoverability can only be disabled.
		{"0s", true, true},
		{"0d", true, true},
		{"0s", false, false},
		{"30d", true, false},
		{"30d", false, false},
		// Values that are still unknown at plan time arrive as the zero value.
		{"", true, false},
	}

	for _, tc := range cases {
		err := validateRetentionPolicyRecoverability(tc.softDeletePeriod, tc.recoverability)
		if gotError := err != nil; gotError != tc.expectError {
			t.Errorf("soft_delete_period %q with recoverability %t: expected error=%t, got %v", tc.softDeletePeriod, tc.recoverability, tc.expectError, err)
		}
	}
}
