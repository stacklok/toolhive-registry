package main

import (
	"strings"
	"testing"
)

func TestValidateBuildFormat(t *testing.T) {
	t.Parallel()

	const legacyRemovedMsg = "legacy registry format was removed"

	tests := []struct {
		name    string
		format  string
		wantErr string
	}{
		{name: "default", format: ""},
		{name: "upstream", format: "upstream"},
		{name: "case insensitive upstream", format: "UpStream"},
		{name: "old all flag value", format: "all"},
		{name: "legacy toolhive", format: "toolhive", wantErr: legacyRemovedMsg},
		{name: "case insensitive legacy toolhive", format: "TOOLHIVE", wantErr: legacyRemovedMsg},
		{name: "legacy alias", format: "legacy", wantErr: legacyRemovedMsg},
		{name: "unknown", format: "json", wantErr: `unknown format "json"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateBuildFormat(tt.format)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateBuildFormat(%q) returned error: %v", tt.format, err)
				}
				return
			}

			if err == nil {
				t.Fatalf("validateBuildFormat(%q) returned nil error, want %q", tt.format, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateBuildFormat(%q) error = %q, want to contain %q", tt.format, err.Error(), tt.wantErr)
			}
		})
	}
}
