package validate

import (
	"regexp"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func StringLengthBetween(min, max int) schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf("expected type of %s to be string", k)
		}

		if len(v) < min || len(v) > max {
			return diag.Errorf("expected length of %s to be in the range (%d - %d), got %s", k, min, max, v)
		}

		return nil
	}
}

func StringInSlice(valid []string) schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf("expected type of %s to be string", k)
		}

		for _, str := range valid {
			if v == str {
				return nil
			}
		}

		return diag.Errorf("expected %s to be one of %v, got %s", k, valid, v)
	}
}

func StringIsNotEmpty(i interface{}, k cty.Path) diag.Diagnostics {
	v, ok := i.(string)
	if !ok {
		return diag.Errorf("expected type of %q to be string", k)
	}

	if v == "" {
		return diag.Errorf("expected %q to not be an empty string, got %v", k, i)
	}

	return nil
}

func StringMatch(r *regexp.Regexp, message string) schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf("expected type of %s to be string", k)
		}

		if ok := r.MatchString(v); !ok {
			if message != "" {
				return diag.Errorf("invalid value for %s (%s)", k, message)

			}
			return diag.Errorf("expected value of %s to match regular expression %q, got %v", k, r, i)
		}
		return nil
	}
}

func StringIsUUID(i interface{}, k cty.Path) diag.Diagnostics {
	v, ok := i.(string)
	if !ok {
		return diag.Errorf("expected type of %q to be string", k)
	}

	if _, err := uuid.ParseUUID(v); err != nil {
		return diag.Errorf("expected %q to be a valid UUID, got %v", k, v)
	}

	return nil
}

func StringIsSystemOrUUID(i interface{}, k cty.Path) diag.Diagnostics {
	v, ok := i.(string)
	if !ok {
		return diag.Errorf("expected type of %q to be string", k)
	}

	if v == "system" {
		return nil
	}

	if _, err := uuid.ParseUUID(v); err != nil {
		return diag.Errorf("expected the value \"system\" or a valid UUID, got: %v", v)
	}

	return nil
}
