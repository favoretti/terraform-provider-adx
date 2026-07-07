package adx

import "testing"

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
