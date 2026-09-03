package utils_test

import (
	"strings"
	"testing"

	"github.com/msrevive/nexus2/internal/payload"
	"github.com/msrevive/nexus2/pkg/utils"
)

func TestProcessJSON(t *testing.T) {
	tests := []struct{
		name string
		body string
		wantErr string // substring the error must contain, "" means no error expected
		check func(*testing.T, payload.Character)
	}{
		{
			name: "valid",
			body: `{"id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","steamid":"76561197960265728","slot":2,"size":5514,"data":"AAA=","flags":3}`,
			check: func(t *testing.T, c payload.Character) {
				if c.SteamID != "76561197960265728" {
					t.Errorf("SteamID = %q, want %q", c.SteamID, "76561197960265728")
				}
				if c.Slot != 2 {
					t.Errorf("Slot = %d, want 2", c.Slot)
				}
				if c.Flags != 3 {
					t.Errorf("Flags = %d, want 3", c.Flags)
				}
				if c.ID.String() != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
					t.Errorf("ID = %q, want 6ba7b810-9dad-11d1-80b4-00c04fd430c8", c.ID)
				}
			},
		},
		{
			name: "syntax error",
			body: `{"steamid":`,
			wantErr: "json syntax error at byte",
		},
		{
			name: "type mismatch",
			body: `{"slot":"abc"}`,
			wantErr: `json type mismatch for field "slot"`,
		},
		{
			name: "empty body",
			body: ``,
			wantErr: "request body is empty",
		},
		{
			name: "whitespace only body",
			body: "  \n\t ",
			wantErr: "request body is empty",
		},
		{
			// The pinned MatchCaseInsensitiveNames option keeps v1 behavior; without it
			// json/v2 would match names case-sensitively and leave SteamID unset.
			name: "case insensitive names",
			body: `{"SteamID":"76561197960265728","SLOT":1}`,
			check: func(t *testing.T, c payload.Character) {
				if c.SteamID != "76561197960265728" {
					t.Errorf("SteamID = %q, want %q", c.SteamID, "76561197960265728")
				}
				if c.Slot != 1 {
					t.Errorf("Slot = %d, want 1", c.Slot)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var char payload.Character
			err := utils.ProcessJSON([]byte(tt.body), &char)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, char)
			}
		})
	}
}
