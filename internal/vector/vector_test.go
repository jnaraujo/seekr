package vector

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDotProduct(t *testing.T) {
	a := []float32{1, 2, 3}
	b := []float32{4, 5, 6}

	result, err := DotProduct(a, b)
	assert.NoError(t, err)
	assert.Equal(t, float32(32), result)

	aShort := []float32{1, 2}
	_, err = DotProduct(aShort, b)
	assert.Error(t, err)
}

func TestSubtract(t *testing.T) {
	a := []float32{5, 7, 9}
	b := []float32{1, 2, 3}

	result, err := Subtract(a, b)
	assert.NoError(t, err)

	assert.Equal(t, []float32{4, 5, 6}, result)

	aShort := []float32{1, 2}
	_, err = Subtract(aShort, b)
	assert.Error(t, err)
}

func TestNormalize(t *testing.T) {
	v := []float32{3, 4}
	normed := Normalize(v)

	magnitude := float64(math.Sqrt(float64(normed[0]*normed[0] + normed[1]*normed[1])))
	assert.InDelta(t, 1.0, magnitude, isNormalizedPrecisionTolerance)

	zero := []float32{0, 0, 0}
	resZero := Normalize(zero)
	assert.Equal(t, zero, resZero)
}

func TestIsNormalized(t *testing.T) {
	normalized := []float32{1 / float32(math.Sqrt(2)), 1 / float32(math.Sqrt(2))}
	assert.True(t, IsNormalized(normalized))

	notNormalized := []float32{2, 0}
	assert.False(t, IsNormalized(notNormalized))

	epsilon := float32(1e-7)
	almost := []float32{1 + epsilon, 0}

	almostNorm := Normalize(almost)
	assert.True(t, IsNormalized(almostNorm))

	outOfTol := []float32{(1 + 2*float32(isNormalizedPrecisionTolerance)), 0}
	assert.False(t, IsNormalized(outOfTol))
}
