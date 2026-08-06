package adx

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestSuppressExternalTableFolderDiff(t *testing.T) {
	cases := []struct {
		name     string
		old      string
		new      string
		suppress bool
	}{
		{"quoted config vs unquoted adx", "CUR/COM", "'CUR/COM'", true},
		{"double quoted", "CUR/COM", "\"CUR/COM\"", true},
		{"identical", "CUR/COM", "CUR/COM", true},
		{"whitespace", "CUR/COM", " 'CUR/COM' ", true},
		{"real change", "CUR/COM", "'NEW/FOLDER'", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := suppressExternalTableFolderDiff("folder", c.old, c.new, nil); got != c.suppress {
				t.Fatalf("suppress(%q,%q) = %v, want %v", c.old, c.new, got, c.suppress)
			}
		})
	}
}

func TestSuppressExternalTablePathFormatDiff(t *testing.T) {
	cases := []struct {
		name     string
		old      string
		new      string
		suppress bool
	}{
		{
			"quote style and spacing",
			`datetime_pattern("yyyy/MM/dd",Day)`,
			`datetime_pattern('yyyy/MM/dd', Day)`,
			true,
		},
		{"identical", `datetime_pattern('yyyy/MM/dd', Day)`, `datetime_pattern('yyyy/MM/dd', Day)`, true},
		{"real change", `datetime_pattern("yyyy/MM/dd",Day)`, `datetime_pattern('yyyy/MM', Day)`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := suppressExternalTablePathFormatDiff("path_format", c.old, c.new, nil); got != c.suppress {
				t.Fatalf("suppress(%q,%q) = %v, want %v", c.old, c.new, got, c.suppress)
			}
		})
	}
}

func TestSuppressExternalTablePartitionsDiff(t *testing.T) {
	cases := []struct {
		name     string
		old      string // as returned by ADX (JSON) or previously stored
		new      string // as written in config (KQL)
		suppress bool
	}{
		{
			"datetime function json vs kql",
			`[{"ColumnName":"ingestion_timestamp","Function":"StartOfDay","Name":"Day","Ordinal":0}]`,
			"Day:datetime = startofday (ingestion_timestamp)",
			true,
		},
		{
			"kql vs kql spacing/case",
			"Day:datetime = startofday(ingestion_timestamp)",
			"Day:datetime = StartOfDay (ingestion_timestamp)",
			true,
		},
		{
			"hash partition json vs kql",
			`[{"ColumnName":"CustomerId","Function":"Hash","Mod":10,"Name":"CustomerIdHash","Ordinal":0}]`,
			"CustomerIdHash:long = hash(CustomerId, 10)",
			true,
		},
		{
			"empty variants",
			"[]",
			"jsonencode([])",
			true,
		},
		{
			"multiple partitions ordered by ordinal",
			`[{"ColumnName":"c1","Function":"StartOfMonth","Name":"Month","Ordinal":1},{"ColumnName":"ts","Function":"StartOfDay","Name":"Day","Ordinal":0}]`,
			"Day:datetime = startofday(ts), Month:datetime = startofmonth(c1)",
			true,
		},
		{
			"real change in function",
			`[{"ColumnName":"ingestion_timestamp","Function":"StartOfDay","Name":"Day","Ordinal":0}]`,
			"Day:datetime = startofmonth (ingestion_timestamp)",
			false,
		},
		{
			"real change in column",
			`[{"ColumnName":"ingestion_timestamp","Function":"StartOfDay","Name":"Day","Ordinal":0}]`,
			"Day:datetime = startofday (event_timestamp)",
			false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := suppressExternalTablePartitionsDiff("partitions", c.old, c.new, nil); got != c.suppress {
				t.Fatalf("suppress(old=%q, new=%q) = %v, want %v", c.old, c.new, got, c.suppress)
			}
		})
	}
}

// The fixtures below drive resourceADXExternalTable through the real SDK diff/apply
// pipeline (not d.Set fakery) to reproduce the exact conditions ADX and Terraform
// produce: partitions/path_format/folder stored in ADX's canonicalized form in state,
// versus the KQL the user writes in config.
const (
	extTableKQLPartitions = "Day:datetime = startofday(ingestion_timestamp)"
	extTableADXPartitions = `[{"ColumnName":"ingestion_timestamp","Function":"StartOfDay","Name":"Day","Ordinal":0}]`
	extTableKQLPathFormat = `datetime_pattern('yyyy/MM/dd', Day)`
	extTableADXPathFormat = `datetime_pattern("yyyy/MM/dd",Day)`
	extTableKQLFolder     = `'CUR/COM'`
	extTableADXFolder     = `CUR/COM`
)

// externalTableRawConfig builds a raw HCL-equivalent config map for the resource,
// applying any overrides on top of a fixed set of base attributes.
func externalTableRawConfig(dataFormat string, overrides map[string]interface{}) map[string]interface{} {
	cfg := map[string]interface{}{
		"name":                      "MyTable",
		"database_name":             "db",
		"data_format":               dataFormat,
		"storage_connection_string": "https://sa.blob.core.windows.net/container;secret",
		"schema":                    "a:string",
		"kind":                      "storage",
		"partitions":                extTableKQLPartitions,
		"path_format":               extTableKQLPathFormat,
		"folder":                    extTableKQLFolder,
		"include_headers":           "All",
		"encoding":                  "UTF8NoBOM",
	}
	for k, v := range overrides {
		cfg[k] = v
	}
	return cfg
}

// externalTableStoredState builds the prior state InstanceState as it would look after
// resourceADXExternalTableRead persisted ADX's canonicalized values.
func externalTableStoredState(dataFormat string, overrides map[string]string) *terraform.InstanceState {
	attrs := map[string]string{
		"name":                      "MyTable",
		"database_name":             "db",
		"data_format":               dataFormat,
		"storage_connection_string": "https://sa.blob.core.windows.net/container;secret",
		"schema":                    "a:string",
		"kind":                      "storage",
		"partitions":                extTableADXPartitions,
		"path_format":               extTableADXPathFormat,
		"folder":                    extTableADXFolder,
		"include_headers":           "All",
		"encoding":                  "UTF8NoBOM",
	}
	for k, v := range overrides {
		attrs[k] = v
	}
	return &terraform.InstanceState{
		ID:         "cluster.example.com|db|externaltable|MyTable",
		Attributes: attrs,
	}
}

// applyExternalTableConfig runs config (and, for updates, priorState) through the
// resource's real Diff/Apply pipeline and returns the *schema.ResourceData that
// Create/UpdateContext received, without ever touching the network: CreateContext and
// UpdateContext are swapped for a capturing stub for the duration of the call.
func applyExternalTableConfig(t *testing.T, priorState *terraform.InstanceState, rawConfig map[string]interface{}) *schema.ResourceData {
	t.Helper()

	r := resourceADXExternalTable()
	meta := &Meta{DefaultClusterConfig: &ClusterConfig{}}
	config := terraform.NewResourceConfigRaw(rawConfig)

	diff, err := r.Diff(context.Background(), priorState, config, meta)
	if err != nil {
		t.Fatalf("Diff() error: %v", err)
	}
	if diff == nil {
		t.Fatal("Diff() returned nil (no changes) but this test expects to apply one")
	}

	// Mirrors what terraform-plugin-sdk's gRPC bridge does at apply time: it attaches
	// the real, fully-resolved config as RawConfig on top of the legacy attribute
	// diff, independent of whatever DiffSuppressFunc did to individual attributes.
	rawVals := map[string]cty.Value{}
	for _, key := range []string{"partitions", "path_format", "folder"} {
		if v, ok := rawConfig[key]; ok {
			rawVals[key] = cty.StringVal(v.(string))
		} else {
			rawVals[key] = cty.NullVal(cty.String)
		}
	}
	diff.RawConfig = cty.ObjectVal(rawVals)

	var captured *schema.ResourceData
	stub := func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
		captured = d
		return nil
	}
	r.CreateContext = stub
	r.UpdateContext = stub

	if _, diags := r.Apply(context.Background(), priorState, diff, meta); diags.HasError() {
		t.Fatalf("Apply() diags: %v", diags)
	}
	if captured == nil {
		t.Fatal("Create/UpdateContext was not invoked")
	}
	return captured
}

func assertExternalTableStatementUsesKQL(t *testing.T, statement, dataFormat string) {
	t.Helper()

	wantPartition := "partition by (" + extTableKQLPartitions + ")"
	wantPathFormat := "pathformat = (" + extTableKQLPathFormat + ")"
	wantFolder := "folder=" + extTableKQLFolder
	wantDataFormat := "dataformat = " + dataFormat

	for _, want := range []string{wantPartition, wantPathFormat, wantFolder, wantDataFormat} {
		if !strings.Contains(statement, want) {
			t.Errorf("statement %q missing %q", statement, want)
		}
	}
	for _, unwanted := range []string{"ColumnName", extTableADXPathFormat, "folder=" + extTableADXFolder} {
		if strings.Contains(statement, unwanted) {
			t.Errorf("statement %q unexpectedly contains stale/canonicalized value %q", statement, unwanted)
		}
	}
}

// TestExternalTableCreateStatementUsesConfiguredKQL covers a brand-new resource: there
// is no prior state to suppress against, so this also pins down that the fix leaves
// Create's behavior unchanged.
func TestExternalTableCreateStatementUsesConfiguredKQL(t *testing.T) {
	d := applyExternalTableConfig(t, nil, externalTableRawConfig("csv", nil))
	assertExternalTableStatementUsesKQL(t, buildExternalTableCreateStatement(d), "csv")
}

// TestExternalTableCreateStatementQuotesDocString reproduces a bug found via
// TestAccADXExternalTable_basic against a live cluster: doc_string was inserted into the
// with() clause unquoted (unlike folder, which existing configs may already quote
// themselves -- see the DiffSuppressFunc tests above -- doc_string has no such
// accommodation and is never read back by Read, so there's no established
// "quote it yourself" convention to preserve). Any doc_string containing a space -- which
// is the entire point of a free-text documentation string -- produced invalid KQL
// ("docString=external table regression test" parses as a bareword followed by
// unexpected tokens). It must come out quoted, matching how adx_table quotes its
// docstring.
func TestExternalTableCreateStatementQuotesDocString(t *testing.T) {
	d := applyExternalTableConfig(t, nil, externalTableRawConfig("csv", map[string]interface{}{
		"doc_string": "external table regression test",
	}))

	statement := buildExternalTableCreateStatement(d)
	if !strings.Contains(statement, "docString='external table regression test'") {
		t.Fatalf("statement %q missing quoted docString", statement)
	}
}

// TestExternalTableUpdateStatementUsesConfiguredKQL reproduces the v0.0.41 regression:
// state holds ADX's canonicalized partitions/path_format/folder, an unrelated field
// (data_format) changes to force a real update, and the DiffSuppressFuncs drop
// partitions/path_format/folder from the diff because they're semantically unchanged.
// The resulting .create-or-alter statement must still carry the user's KQL, not the
// stale canonicalized state.
func TestExternalTableUpdateStatementUsesConfiguredKQL(t *testing.T) {
	priorState := externalTableStoredState("csv", nil)
	rawConfig := externalTableRawConfig("json", nil) // unrelated change forces the update

	d := applyExternalTableConfig(t, priorState, rawConfig)

	// Sanity check on the premise: d.Get must still reflect the stale, suppressed
	// state, otherwise this test would pass even without the fix.
	if got := d.Get("partitions").(string); got != extTableADXPartitions {
		t.Fatalf("premise broken: expected d.Get(partitions) to be the stale state JSON %q, got %q", extTableADXPartitions, got)
	}

	assertExternalTableStatementUsesKQL(t, buildExternalTableCreateStatement(d), "json")
}

// TestExternalTableNoChangesProducesNoDiff covers an existing resource whose config
// hasn't changed since the last apply: state already holds ADX's canonicalized
// partitions/path_format/folder for that exact KQL config. Diff must come back empty
// (nil) so Terraform reports no changes and never calls Update at all -- proving the
// fix didn't reintroduce the permanent diff that v0.0.41 was written to eliminate.
func TestExternalTableNoChangesProducesNoDiff(t *testing.T) {
	r := resourceADXExternalTable()
	meta := &Meta{DefaultClusterConfig: &ClusterConfig{}}

	priorState := externalTableStoredState("csv", nil)
	config := terraform.NewResourceConfigRaw(externalTableRawConfig("csv", nil))

	diff, err := r.Diff(context.Background(), priorState, config, meta)
	if err != nil {
		t.Fatalf("Diff() error: %v", err)
	}
	if diff != nil {
		t.Fatalf("expected no diff for an unchanged resource, got %+v", diff.Attributes)
	}
}

// testAccExternalTableStorageConnectionString returns the storage connection string used
// by the adx_external_table acceptance tests below. Contrary to what the docs on
// validate_not_empty suggest, ADX resolves the storage hostname via DNS as part of
// .create-or-alter external table even without that flag set (confirmed against a live
// cluster: a syntactically valid but non-existent hostname is rejected with
// RequestCalloutDnsCheckFailedException), so a placeholder default can't work here. The
// test is skipped, not failed, when this isn't configured, since it requires standing up
// real storage that's independent of the ADX cluster itself.
func testAccExternalTableStorageConnectionString(t *testing.T) string {
	t.Helper()
	v := os.Getenv("ADX_TEST_STORAGE_CONNECTION_STRING")
	if v == "" {
		t.Skip("ADX_TEST_STORAGE_CONNECTION_STRING must be set to a real storage container connection string (e.g. \"https://<account>.blob.core.windows.net/<container>;<SAS token or account key>\") -- ADX resolves the storage hostname via DNS at external table creation time, so no placeholder value works")
	}
	return v
}

const (
	accExtTableKQLPartitions = "Day:datetime = startofday(ingestion_timestamp)"
	accExtTableKQLPathFormat = `datetime_pattern('yyyy/MM/dd', Day)`
	accExtTableKQLFolder     = "'extacctestfolder'"
	accExtTableSchema        = "ingestion_timestamp:datetime,name:string,value:int"
)

type ADXExternalTableTestResource struct{}

// TestAccADXExternalTable_basic is the acceptance-level regression test for the v0.0.41
// bug: partitions/path_format/folder are configured in KQL, ADX stores them back in its
// own canonicalized form, and a second step changes an unrelated field (doc_string) to
// force a real Update while that canonicalized state is still on-record. Before the fix,
// that Update sent ADX's canonicalized values back to it and failed outright; the
// resource.Test framework additionally re-plans after every step and fails on a non-empty
// diff, which also catches the permanent-diff half of the original v0.0.41 bug.
func TestAccADXExternalTable_basic(t *testing.T) {
	storageConnectionString := testAccExternalTableStorageConnectionString(t)

	var entity ADXExternalTable
	r := ADXExternalTableTestResource{}
	rtcBuilder := BuildResourceTestContext[ADXExternalTable]()
	rtc, _ := rtcBuilder.Test(t).Type("adx_external_table").
		DatabaseName(testAccDatabaseName()).
		EntityType("external table").
		ReadStatementFunc(func(id string) (string, error) {
			extId, err := parseADXExternalTableID(id)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(".show external table %s ", extId.Name), nil
		}).Build()

	// ADX reports partitions/path_format/folder back in its own canonicalized form (see
	// the DiffSuppressFuncs above), not the KQL that was configured, so state-level
	// TestCheckResourceAttr can't assert the KQL literal. Instead check what ADX itself
	// parsed out of the KQL we sent it, via the entity GetTestCheckEntityExists reads.
	checkADXGotThePartitionScheme := func(*terraform.State) error {
		if !strings.Contains(strings.ToLower(entity.Partitions), "startofday") {
			return fmt.Errorf("expected ADX to report partitions reflecting startofday(), got %q", entity.Partitions)
		}
		if entity.Folder == "" {
			return fmt.Errorf("expected ADX to report a non-empty folder")
		}
		if entity.PathFormat == "" {
			return fmt.Errorf("expected ADX to report a non-empty path_format")
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: rtc.GetTestCheckEntityDestroyed(),
		Steps: []resource.TestStep{
			{
				Config: r.basic(rtc, storageConnectionString),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "name", rtc.EntityName),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "database_name", rtc.DatabaseName),
					checkADXGotThePartitionScheme,
				),
			},
			{
				// Same partitions/path_format/folder KQL as step 1, but doc_string is now
				// set: this forces a real Update while ADX's canonicalized state for the
				// other three fields is suppressed in the diff -- the exact v0.0.41
				// regression, where the resulting management command used ADX's
				// canonicalized (non-KQL) values and was rejected.
				Config: r.basic_update(rtc, storageConnectionString),
				Check: resource.ComposeTestCheckFunc(
					rtc.GetTestCheckEntityExists(&entity),
					resource.TestCheckResourceAttr(rtc.GetTFName(), "doc_string", "external table regression test"),
					checkADXGotThePartitionScheme,
				),
			},
		},
	})
}

func (this ADXExternalTableTestResource) basic(rtc *ResourceTestContext[ADXExternalTable], storageConnectionString string) string {
	return fmt.Sprintf(`

	resource "%s" %s {
		database_name              = "%s"
		name                       = "%s"
		data_format                = "csv"
		storage_connection_string  = "%s"
		schema                     = "%s"
		partitions                 = "%s"
		path_format                = "%s"
		folder                     = "%s"
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName, storageConnectionString, accExtTableSchema, accExtTableKQLPartitions, accExtTableKQLPathFormat, accExtTableKQLFolder)
}

func (this ADXExternalTableTestResource) basic_update(rtc *ResourceTestContext[ADXExternalTable], storageConnectionString string) string {
	return fmt.Sprintf(`

	resource "%s" %s {
		database_name              = "%s"
		name                       = "%s"
		data_format                = "csv"
		storage_connection_string  = "%s"
		schema                     = "%s"
		partitions                 = "%s"
		path_format                = "%s"
		folder                     = "%s"
		doc_string                 = "external table regression test"
	}
	`, rtc.Type, rtc.Label, rtc.DatabaseName, rtc.EntityName, storageConnectionString, accExtTableSchema, accExtTableKQLPartitions, accExtTableKQLPathFormat, accExtTableKQLFolder)
}
