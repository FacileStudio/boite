package qemu

import (
	"testing"
	"time"
)

func TestExpiryMarker(t *testing.T) {
	if got := expiryMarker("CASIER_TOKEN"); got != "CASIER_TOKEN__expires" {
		t.Fatalf("expiryMarker = %q, want %q", got, "CASIER_TOKEN__expires")
	}
}

func TestParseExpiryMarker(t *testing.T) {
	got, err := parseExpiryMarker("1768000000")
	if err != nil {
		t.Fatalf("parseExpiryMarker: %v", err)
	}
	if got.Unix() != 1768000000 {
		t.Fatalf("parseExpiryMarker = %d, want 1768000000", got.Unix())
	}
	padded, err := parseExpiryMarker(" 1768000000 \n")
	if err != nil || padded.Unix() != 1768000000 {
		t.Fatalf("parseExpiryMarker padded = %d, %v", padded.Unix(), err)
	}
	if _, err := parseExpiryMarker("not-a-number"); err == nil {
		t.Fatal("expected a garbage marker to fail parsing")
	}
}

func TestClassifyCasierToken(t *testing.T) {
	now := time.Unix(2000, 0)
	cases := []struct {
		name    string
		present bool
		marker  string
		want    casierTokenState
	}{
		{"absent key", false, "3000", casierTokenAbsent},
		{"missing marker", true, "", casierTokenUnmarked},
		{"garbage marker", true, "soon-ish", casierTokenUnmarked},
		{"fresh token", true, "3000", casierTokenFresh},
		{"expired token", true, "1999", casierTokenExpired},
		{"boundary is expired", true, "2000", casierTokenExpired},
	}
	for _, tc := range cases {
		got := classifyCasierToken(tc.present, tc.marker, now)
		if got != tc.want {
			t.Errorf("%s: classifyCasierToken = %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestParseCasierProbe(t *testing.T) {
	present := parseCasierProbe("k=yes\nexp=1768000000\n")
	if !present.KeyPresent || present.Marker != "1768000000" {
		t.Fatalf("parseCasierProbe present = %+v", present)
	}
	absent := parseCasierProbe("k=no\nexp=\n")
	if absent.KeyPresent || absent.Marker != "" {
		t.Fatalf("parseCasierProbe absent = %+v", absent)
	}
}

func TestCasierTokenTTL(t *testing.T) {
	cases := []struct {
		ttl     string
		want    time.Duration
		wantErr bool
	}{
		{"", 0, false},
		{"8h", 8 * time.Hour, false},
		{"30m", 30 * time.Minute, false},
		{"nope", 0, true},
		{"0s", 0, true},
		{"-1h", 0, true},
	}
	for _, tc := range cases {
		c := CasierEnv{TTL: tc.ttl}
		got, err := c.tokenTTL()
		if tc.wantErr {
			if err == nil {
				t.Errorf("ttl %q: expected an error", tc.ttl)
			}
			continue
		}
		if err != nil {
			t.Errorf("ttl %q: %v", tc.ttl, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ttl %q = %v, want %v", tc.ttl, got, tc.want)
		}
	}
}
