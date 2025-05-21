package vector

import (
	"errors"
	"math"
)

const isNormalizedPrecisionTolerance = 1e-6

func DotProduct(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, errors.New("vectors must have the same length")
	}
	var sum float32
	for i, v := range a {
		sum += v * b[i]
	}
	return sum, nil
}

func Normalize(v []float32) []float32 {
	var sumSquares float64
	for _, x := range v {
		sumSquares += float64(x * x)
	}
	norm := float32(math.Sqrt(sumSquares))
	if norm == 0 {
		return v
	}

	res := make([]float32, len(v))
	for i, x := range v {
		res[i] = x / norm
	}
	return res
}

func Subtract(a, b []float32) ([]float32, error) {
	if len(a) != len(b) {
		return nil, errors.New("vectors must have the same length")
	}

	res := make([]float32, len(a))
	for i, val := range a {
		res[i] = val - b[i]
	}

	return res, nil
}

func IsNormalized(v []float32) bool {
	var sumSquares float64
	for _, x := range v {
		sumSquares += float64(x) * float64(x)
	}

	magnitude := math.Sqrt(sumSquares)
	return math.Abs(magnitude-1) < isNormalizedPrecisionTolerance
}

func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float32
	for i, v := range a {
		dot += v * b[i]
		normA += v * v
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / float32(math.Sqrt(float64(normA*normB)))
}

// FastCosineSimilarity assumes that both a and b are unit vectors.
// For normalized vectors, cosine similarity reduces to their dot product.
func FastCosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	var dot float32
	for i, v := range a {
		dot += v * b[i]
	}
	return dot
}
