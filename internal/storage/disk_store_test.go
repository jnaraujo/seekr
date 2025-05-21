package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jnaraujo/seekr/internal/document"
	"github.com/jnaraujo/seekr/internal/vector"
	"github.com/stretchr/testify/assert"
)

func makeTempStore(t *testing.T) (*DiskStore, func()) {
	dir := t.TempDir()
	file := filepath.Join(dir, "store.skdb")
	ds, err := NewDiskStore(file)
	assert.NoError(t, err)

	return ds, func() {
		ds.file.Close()
		os.Remove(file)
	}
}

func TestIndexAndGet(t *testing.T) {
	ds, cleanup := makeTempStore(t)
	defer cleanup()

	ctx := context.Background()
	doc, err := document.NewDocument("doc1", nil, []float32{1.0, 0.0}, "empty")
	assert.NoError(t, err)

	err = ds.Index(ctx, doc)
	assert.NoError(t, err)

	got, err := ds.Get(ctx, "doc1")
	assert.NoError(t, err)
	assert.Equal(t, doc, got)

	_, err = ds.Get(ctx, "random-id")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestSearchOrdering(t *testing.T) {
	ds, cleanup := makeTempStore(t)
	defer cleanup()

	ctx := context.Background()

	d1, _ := document.NewDocument("a", nil, []float32{1.0, 0.0}, "")
	d2, _ := document.NewDocument("b", nil, []float32{0.0, 1.0}, "")

	assert.NoError(t, ds.Index(ctx, d1))
	assert.NoError(t, ds.Index(ctx, d2))

	query := []float32{0.1, 0.9}

	results, err := ds.Search(ctx, query, 2)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	assert.Equal(t, "b", results[0].Document.ID)
	assert.True(t, vector.CosineSimilarity(query, d2.Embedding) >= results[1].Score)
}

func TestSearchEmpty(t *testing.T) {
	ds, cleanup := makeTempStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := ds.Search(ctx, []float32{1, 2, 3}, 1)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestPersistenceAcrossLoads(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "store.jsonl")

	{ // first load
		ds, err := NewDiskStore(file)
		assert.NoError(t, err)
		defer ds.file.Close()

		doc, _ := document.NewDocument("persist", nil, []float32{0.5, 0.5}, "content")
		assert.NoError(t, ds.Index(context.Background(), doc))
	}

	{ // second load
		ds2, err := NewDiskStore(file)
		assert.NoError(t, err)
		defer ds2.file.Close()

		got, err := ds2.Get(context.Background(), "persist")
		assert.NoError(t, err)
		assert.Equal(t, "persist", got.ID)

		// Also ensure that the file contains a correct JSON line
		bytes, err := os.ReadFile(file)
		assert.NoError(t, err)
		var docs []document.Document
		for _, line := range splitLines(bytes) {
			var d document.Document
			assert.NoError(t, json.Unmarshal(line, &d))
			docs = append(docs, d)
		}
		assert.Len(t, docs, 1)
		assert.Equal(t, "persist", docs[0].ID)
	}
}

func splitLines(b []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, c := range b {
		if c == '\n' {
			lines = append(lines, b[start:i])
			start = i + 1
		}
	}
	if start < len(b) {
		lines = append(lines, b[start:])
	}
	return lines
}
