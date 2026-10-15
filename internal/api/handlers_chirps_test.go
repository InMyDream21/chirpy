package api

import "testing"

func TestCleanBody(t *testing.T) {
	tests := map[string]struct{ in, want string }{
		"no forbidden words": {"hello world", "hello world"},
		"replaces words":     {"This is a kerfuffle opinion", "This is a **** opinion"},
		"case insensitive":   {"Sharbert FORNAX", "**** ****"},
		"punctuation kept":   {"Sharbert!", "Sharbert!"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cleanBody(tc.in); got != tc.want {
				t.Errorf("cleanBody(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
