package money

import "testing"

func TestParse(t *testing.T) {
	tests := map[string]int64{
		"12.34": 1234,
		"12":    1200,
		"0.5":   50,
		"-7.01": -701,
		"+3.20": 320,
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := Parse(input)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("Parse(%q) = %d, want %d", input, got, want)
			}
		})
	}
}

func TestParseRejectsFractionsOfACent(t *testing.T) {
	if _, err := Parse("1.001"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestFormat(t *testing.T) {
	if got := Format(-701); got != "-7.01" {
		t.Fatalf("Format(-701) = %q", got)
	}
}
