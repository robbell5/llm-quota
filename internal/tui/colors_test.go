package tui

import (
	"testing"

	"github.com/lucasb-eyer/go-colorful"
)

func nearHex(t *testing.T, got colorful.Color, hex string) {
	t.Helper()
	want, err := colorful.Hex(hex)
	if err != nil {
		t.Fatalf("bad ref hex %q: %v", hex, err)
	}
	if got.DistanceLab(want) > 0.02 {
		t.Fatalf("color %v not near %s (%v)", got.Hex(), hex, want.Hex())
	}
}

func TestRampColorfulEndpoints(t *testing.T) {
	nearHex(t, rampColorful(0), "#a6e3a1") // green at empty
	nearHex(t, rampColorful(1), "#f38ba8") // red at full
}

func TestRampColorfulMidrangeIsAmberNotGray(t *testing.T) {
	_, chroma, _ := rampColorful(0.5).Hcl()
	if chroma < 0.15 {
		t.Fatalf("midrange should be chromatic amber, got chroma %v", chroma)
	}
}

func TestRampColorClampsOutOfRange(t *testing.T) {
	nearHex(t, rampColorful(-0.5), "#a6e3a1")
	nearHex(t, rampColorful(1.5), "#f38ba8")
}
