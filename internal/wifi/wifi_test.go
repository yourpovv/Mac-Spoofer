package wifi

import "testing"

func TestNormalizeMac(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{"colons", "02:1A:2B:3C:4D:5E", "021A2B3C4D5E", false},
		{"dashes lowercase", "0e-1a-2b-3c-4d-5e", "0E1A2B3C4D5E", false},
		{"bare hex", "121A2B3C4D5E", "121A2B3C4D5E", false},
		{"too short", "02:1A", "", true},
		{"multicast first byte", "03:1A:2B:3C:4D:5E", "", true},
		{"globally unique burned-in", "6C:CD:D6:C0:DC:2A", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeMac(tt.raw)
			if tt.wantErr && err == nil {
				t.Fatalf("NormalizeMac(%q) = %q, want error", tt.raw, got)
			}
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("NormalizeMac(%q) error: %v", tt.raw, err)
				}
				if got != tt.want {
					t.Errorf("NormalizeMac(%q) = %q, want %q", tt.raw, got, tt.want)
				}
			}
		})
	}
}

func TestDisplayMac(t *testing.T) {
	if got := DisplayMac("021A2B3C4D5E"); got != "02-1A-2B-3C-4D-5E" {
		t.Errorf("DisplayMac = %q, want %q", got, "02-1A-2B-3C-4D-5E")
	}
}

func TestRandomMacIsSpoofable(t *testing.T) {
	got, err := RandomMac()
	if err != nil {
		t.Fatalf("RandomMac error: %v", err)
	}
	if !isSpoofable(got) {
		t.Errorf("RandomMac = %q, not locally-administered unicast", got)
	}
}
