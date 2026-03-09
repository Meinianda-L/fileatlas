package core

import "testing"

func TestNormalizeShareMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "default full", input: "", want: "full"},
		{name: "private", input: "private", want: "private"},
		{name: "summary", input: "summary", want: "summary"},
		{name: "full uppercase", input: "FULL", want: "full"},
		{name: "trimmed value", input: " summary ", want: "summary"},
		{name: "invalid", input: "public", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := normalizeShareMode(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none (value=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("normalizeShareMode(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
