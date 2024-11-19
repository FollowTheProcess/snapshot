package colour_test

import (
	"testing"

	"github.com/FollowTheProcess/snapshot/internal/colour"
)

func TestColour(t *testing.T) {
	tests := []struct {
		name string // Name of the test case
		got  string // Text to print
		want string // Expected output
	}{
		{
			name: "header",
			got:  colour.Header("header"),
			want: "\x1b[1;0036mheader\x1b[000000m",
		},
		{
			name: "green",
			got:  colour.Green("green"),
			want: "\x1b[0;0032mgreen\x1b[000000m",
		},
		{
			name: "red",
			got:  colour.Red("red"),
			want: "\x1b[0;0031mred\x1b[000000m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got != want; got = %q, want = %q", tt.got, tt.want)
			}
		})
	}
}
