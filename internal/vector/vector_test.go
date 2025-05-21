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

func TestCosineSimilarity(t *testing.T) {
	// identical vectors -> similarity 1
	a := []float32{1, 0, -1}
	b := []float32{1, 0, -1}
	assert.InDelta(t, 1.0, float64(CosineSimilarity(a, b)), 1e-6)

	// orthogonal vectors -> similarity 0
	a = []float32{1, 0}
	b = []float32{0, 1}
	assert.InDelta(t, 0.0, float64(CosineSimilarity(a, b)), 1e-6)

	// length mismatch -> 0
	aShort := []float32{1, 2}
	bLong := []float32{1, 2, 3}
	assert.Equal(t, float32(0), CosineSimilarity(aShort, bLong))

	// zero vector -> 0
	zero := []float32{0, 0, 0}
	assert.Equal(t, float32(0), CosineSimilarity(a, zero))
}

// TestFastCosineSimilarity tests the fast cosine similarity assuming normalized inputs.
func TestFastCosineSimilarity(t *testing.T) {
	// normalized identical vectors -> 1
	a := []float32{1, 0}
	b := []float32{1, 0}
	assert.Equal(t, float32(1), FastCosineSimilarity(a, b))

	// normalized orthogonal vectors -> 0
	a = Normalize([]float32{1, 1})
	b = Normalize([]float32{1, -1})
	assert.InDelta(t, 0.0, float64(FastCosineSimilarity(a, b)), 1e-6)

	// length mismatch -> 0
	aShort := []float32{1}
	assert.Equal(t, float32(0), FastCosineSimilarity(aShort, b))
}
