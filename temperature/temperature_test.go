package temperature

import "testing"

func TestKToC(t *testing.T) {
	if got := KToC(0); got != -273.15 {
		t.Errorf("with K=0, want -273.15, got %.2f", got)
	}
	if got := KToC(273.15); got != 0 {
		t.Errorf("with K=273.15, want 0, got %.2f", got)
	}
}

func TestKToF(t *testing.T) {
	if got := KToF(273.15); got != 32 {
		t.Errorf("with K=273.15, want 32, got %.2f", got)
	}
	if got := KToF(373.15); got != 212 {
		t.Errorf("with K=373.15, want 212, got %.2f", got)
	}
}

func TestSteinhartTemp(t *testing.T) {
	a, b, c := 2.108508173e-3, 0.7979204727e-4, 6.535076315e-7
	if got := KToC(SteinhartTemp(a, b, c, 10000)); !floatEquals(25, got) {
		t.Errorf("with r=10000, want Steinhart temp 25C, got %.6fC", got)
	}
}

const floatEpsilon float64 = 0.000001

func floatEquals(a, b float64) bool {
	if (a-b) < floatEpsilon && (b-a) < floatEpsilon {
		return true
	}
	return false
}
