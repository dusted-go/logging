package httplogger

import (
	"reflect"
	"testing"
)

func TestParseForwarded(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   []ForwardedElement
	}{
		{
			name:   "empty header",
			header: "",
			want:   nil,
		},
		{
			name:   "single element",
			header: "for=192.0.2.60;proto=http;by=203.0.113.43",
			want: []ForwardedElement{
				{
					For:   "192.0.2.60",
					Proto: "http",
					By:    "203.0.113.43",
				},
			},
		},
		{
			name:   "multiple elements",
			header: "for=192.0.2.43, for=198.51.100.17",
			want: []ForwardedElement{
				{
					For: "192.0.2.43",
				},
				{
					For: "198.51.100.17",
				},
			},
		},
		{
			name:   "quoted strings",
			header: `for="192.0.2.60";proto=http;by="203.0.113.43"`,
			want: []ForwardedElement{
				{
					For:   "192.0.2.60",
					Proto: "http",
					By:    "203.0.113.43",
				},
			},
		},
		{
			name:   "case insensitive keys",
			header: "For=192.0.2.60;PROTO=http;By=203.0.113.43",
			want: []ForwardedElement{
				{
					For:   "192.0.2.60",
					Proto: "http",
					By:    "203.0.113.43",
				},
			},
		},
		{
			name:   "ipv6",
			header: `for="[2001:db8:cafe::17]:4711"`,
			want: []ForwardedElement{
				{
					For: "[2001:db8:cafe::17]:4711",
				},
			},
		},
		{
			name:   "multiple parameters with host",
			header: "for=192.0.2.60;proto=http;host=example.com",
			want: []ForwardedElement{
				{
					For:   "192.0.2.60",
					Proto: "http",
					Host:  "example.com",
				},
			},
		},
		{
			name:   "malformed - missing value",
			header: "for=;proto=http",
			want: []ForwardedElement{
				{
					For:   "",
					Proto: "http",
				},
			},
		},
		{
			name:   "malformed - no equals",
			header: "for192.0.2.60;proto=http",
			want: []ForwardedElement{
				{
					Proto: "http",
				},
			},
		},
		{
			name:   "whitespace handling",
			header: " for = 192.0.2.60 ; proto = http ",
			want: []ForwardedElement{
				{
					For:   "192.0.2.60",
					Proto: "http",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseForwarded(tt.header)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseForwarded() = %v, want %v", got, tt.want)
			}
		})
	}
}
