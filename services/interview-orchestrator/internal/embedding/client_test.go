package embedding

import (
	"context"
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	v1 := []float32{1, 0, 0}
	v2 := []float32{1, 0, 0}
	v3 := []float32{0, 1, 0}
	v4 := []float32{-1, 0, 0}

	simSame := CosineSimilarity(v1, v2)
	if math.Abs(simSame-1.0) > 1e-6 {
		t.Errorf("expected 1.0 for identical vectors, got %f", simSame)
	}

	simOrtho := CosineSimilarity(v1, v3)
	if math.Abs(simOrtho-0.0) > 1e-6 {
		t.Errorf("expected 0.0 for orthogonal vectors, got %f", simOrtho)
	}

	simOpp := CosineSimilarity(v1, v4)
	if math.Abs(simOpp - -1.0) > 1e-6 {
		t.Errorf("expected -1.0 for opposite vectors, got %f", simOpp)
	}
}

func TestNormalize(t *testing.T) {
	v := []float32{3, 4}
	normed := Normalize(v)
	var sum float64
	for _, x := range normed {
		sum += float64(x * x)
	}
	length := math.Sqrt(sum)
	if math.Abs(length-1.0) > 1e-6 {
		t.Errorf("expected normalized vector length 1.0, got %f", length)
	}
}

func TestDemoEmbeddingClient(t *testing.T) {
	client := DemoEmbeddingClient{}
	vec, err := client.Embed(context.Background(), "test text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vec) != DefaultDims {
		t.Fatalf("expected length %d, got %d", DefaultDims, len(vec))
	}
	var sum float64
	for _, x := range vec {
		sum += float64(x * x)
	}
	length := math.Sqrt(sum)
	if math.Abs(length-1.0) > 1e-5 {
		t.Errorf("expected unit vector length ~1.0, got %f", length)
	}
}
