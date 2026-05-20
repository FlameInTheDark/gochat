package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompareYugabyteSnapshots(t *testing.T) {
	tests := []struct {
		name      string
		checksum  bool
		source    map[string]yugabyteTableSnapshot
		target    map[string]yugabyteTableSnapshot
		wantOK    bool
		wantTable string
		wantCause string
	}{
		{
			name:     "matching counts and checksums",
			checksum: true,
			source: map[string]yugabyteTableSnapshot{
				"users": {Table: "users", Rows: 2, Checksum: "abc"},
			},
			target: map[string]yugabyteTableSnapshot{
				"users": {Table: "users", Rows: 2, Checksum: "abc"},
			},
			wantOK: true,
		},
		{
			name:     "missing target table",
			checksum: true,
			source: map[string]yugabyteTableSnapshot{
				"guilds": {Table: "guilds", Rows: 1, Checksum: "abc"},
			},
			target:    map[string]yugabyteTableSnapshot{},
			wantOK:    false,
			wantTable: "guilds",
			wantCause: "missing from target",
		},
		{
			name:     "row count mismatch",
			checksum: true,
			source: map[string]yugabyteTableSnapshot{
				"members": {Table: "members", Rows: 3, Checksum: "abc"},
			},
			target: map[string]yugabyteTableSnapshot{
				"members": {Table: "members", Rows: 2, Checksum: "abc"},
			},
			wantOK:    false,
			wantTable: "members",
			wantCause: "row count differs",
		},
		{
			name:     "checksum mismatch",
			checksum: true,
			source: map[string]yugabyteTableSnapshot{
				"roles": {Table: "roles", Rows: 4, Checksum: "abc"},
			},
			target: map[string]yugabyteTableSnapshot{
				"roles": {Table: "roles", Rows: 4, Checksum: "def"},
			},
			wantOK:    false,
			wantTable: "roles",
			wantCause: "checksum differs",
		},
		{
			name:     "checksum mismatch ignored when disabled",
			checksum: false,
			source: map[string]yugabyteTableSnapshot{
				"roles": {Table: "roles", Rows: 4, Checksum: "abc"},
			},
			target: map[string]yugabyteTableSnapshot{
				"roles": {Table: "roles", Rows: 4, Checksum: "def"},
			},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareYugabyteSnapshots("public", tt.checksum, tt.source, tt.target)
			if got.OK != tt.wantOK {
				t.Fatalf("OK = %v, want %v; result=%#v", got.OK, tt.wantOK, got)
			}
			if tt.wantCause == "" {
				return
			}
			for _, table := range got.Tables {
				if table.Table == tt.wantTable {
					if table.Reason != tt.wantCause {
						t.Fatalf("Reason = %q, want %q", table.Reason, tt.wantCause)
					}
					return
				}
			}
			t.Fatalf("table %q not found in result %#v", tt.wantTable, got.Tables)
		})
	}
}

func TestPrintYugabyteVerifyResultText(t *testing.T) {
	result := yugabyteVerifyResult{
		Schema:   "public",
		Checksum: true,
		OK:       false,
		Tables: []yugabyteTableComparison{
			{Table: "users", SourceRows: 1, TargetRows: 2, Status: "mismatch", Reason: "row count differs"},
		},
	}

	var buf bytes.Buffer
	if err := printYugabyteVerifyResult(&buf, "text", result); err != nil {
		t.Fatalf("printYugabyteVerifyResult: %v", err)
	}
	output := buf.String()
	for _, want := range []string{"verification=FAILED", "schema=public", "users", "row count differs"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q missing %q", output, want)
		}
	}
}

func TestQuoteYSQLIdent(t *testing.T) {
	got := quoteYSQLIdent(`weird"name`)
	want := `"weird""name"`
	if got != want {
		t.Fatalf("quoteYSQLIdent() = %q, want %q", got, want)
	}
}
