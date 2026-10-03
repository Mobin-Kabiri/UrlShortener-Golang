package internal

import (
	"errors"
	"math"
	"testing"
)

func TestEncoding(t *testing.T) {
	testsTable := []struct {
		name           string
		input          uint64
		expectedOutput string
		wantErr        error
	}{
		{name: "zero", input: 0, expectedOutput: "0"},
		{name: "lastChar", input: 61, expectedOutput: "Z"},
		{name: "checking above 61", input: 62, expectedOutput: "10"},
		{name: "len=6 output", input: 916132832, expectedOutput: "100000"},
		{name: "last len=5 output", input: 916132831, expectedOutput: "ZZZZZ"},
		{name: "maxInput", input: math.MaxUint32, wantErr: ErrMaxUrlReached},
		{name: "above maxInput", input: math.MaxUint32 + 1, expectedOutput: "0"},
	}

	for _, tt := range testsTable {
		t.Run(tt.name, func(t *testing.T) {
			res, err := EncodeToBase62String(uint32(tt.input))

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("encode(%d) error = %v, want %v", tt.input, err, tt.wantErr)
			}
			if res != tt.expectedOutput {
				t.Errorf("encode(%d) = %q, want %q", tt.input, res, tt.expectedOutput)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	testsTable := []struct {
		name           string
		input          string
		expectedOutput string
		wantErr        error
	}{
		{name: "no scheme", input: "test.com", wantErr: ErrInvalidProtocol},
		{name: "ftp scheme", input: "ftp://test.com", wantErr: ErrInvalidProtocol},
		{name: "typo scheme", input: "htp://test.com", wantErr: ErrInvalidProtocol},
		{name: "single slash", input: "http:/test.com", wantErr: ErrEmptyHost},
		{name: "handling slash", input: "http://test.com/doc/", expectedOutput: "http://test.com/doc"},
		{name: "empty host", input: "http://", wantErr: ErrEmptyHost},
		{name: "plain text", input: "salam bye", wantErr: ErrInvalidProtocol},
		{name: "bad brackets", input: "http://1.2[]:2020", wantErr: ErrBadFormat},
		{name: "bad char", input: "http://sal^^am.com/", wantErr: ErrBadFormat},
		{name: "bad port", input: "http://sal:am.com/", wantErr: ErrBadFormat},
		{name: "empty input", input: "", wantErr: ErrEmptyInput},
		{name: "spaces only", input: "    ", wantErr: ErrEmptyInput},
		{name: "lowercase host", input: "http://sAlam.Com", expectedOutput: "http://salam.com/"},
		{name: "trim spaces", input: "     http://sAlam.Com ", expectedOutput: "http://salam.com/"},

		{name: "http 8080 kept", input: "http://salam.com:8080", expectedOutput: "http://salam.com:8080/"},
		{name: "http 80 removed", input: "http://salam.com:80", expectedOutput: "http://salam.com/"},
		{name: "https 80 kept", input: "https://salam.com:80", expectedOutput: "https://salam.com:80/"},
		{name: "https 8080 kept", input: "https://salam.com:8080", expectedOutput: "https://salam.com:8080/"},
		{name: "https 443 removed", input: "https://salam.com:443", expectedOutput: "https://salam.com/"},
		{name: "http 443 kept", input: "http://salam.com:443", expectedOutput: "http://salam.com:443/"},

		{name: "lowercase scheme", input: "htTp://sAlam.Com/", expectedOutput: "http://salam.com/"},
		{name: "path case kept", input: "htTp://sAlam.Com/AabB", expectedOutput: "http://salam.com/AabB"},
		{name: "extra slashes", input: "http://sAlam.Com//////", expectedOutput: "http://salam.com/"},
		{name: "query no value", input: "https://salam.com/?x", expectedOutput: "https://salam.com/?x="},
		{name: "path slash with query", input: "https://salam.com/doc/?x=1", expectedOutput: "https://salam.com/doc?x=1"},
		{name: "query sorted", input: "http://salam.com/?b=2&a=1", expectedOutput: "http://salam.com/?a=1&b=2"},
		{name: "query sorted no slash", input: "http://salam.com?b=2&a=1", expectedOutput: "http://salam.com/?a=1&b=2"},
		{name: "query sorted dup", input: "http://salam.com/?b=2&a=1", expectedOutput: "http://salam.com/?a=1&b=2"},
		{name: "query fragment removed", input: "http://salam.com/?b=2&a=1#part1", expectedOutput: "http://salam.com/?a=1&b=2"},
		{name: "fragment removed", input: "http://salam.com#part1", expectedOutput: "http://salam.com/"},
	}

	for _, tt := range testsTable {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Normalize(tt.input)
			res2, _ := Normalize(res)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("normal(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
			if res != tt.expectedOutput {
				t.Errorf("normal(%q) = %q, want %q", tt.input, res, tt.expectedOutput)
			}
			if res != res2 {
				t.Fatalf("idempotency is not satisfied")
			}
		})
	}
}

func TestNormalizeOnSameResource(t *testing.T) {
	tests := [...]string{"http://salam.com:80", "http://salam.com/", "http://salam.com", "http://salam.cOm:80"}
	normalizedFirst, err := Normalize(tests[0])
	if err != nil {
		t.Error("unexpected error in normalize function")
	}
	for _, tt := range tests {
		curr, err := Normalize(tt)
		if err != nil {
			t.Error("unexpected error in normalize function")
		}
		if curr != normalizedFirst {
			t.Errorf("normal(%q) = %q, want %q", tt, curr, normalizedFirst)
		}
	}
}
