package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jnaraujo/seekr/internal/document"
	"github.com/jnaraujo/seekr/internal/embeddings"
	"github.com/jnaraujo/seekr/internal/vector"
	"github.com/stretchr/testify/assert"
)

func makeTempStore(t *testing.T) (*DiskStore, func()) {
	dir := t.TempDir()
	file := filepath.Join(dir, "store.skdb")
	ds, err := NewDiskStore(file)
	assert.NoError(t, err)

	return ds, func() {
		ds.Close()
		os.Remove(file)
	}
}

func TestIndexAndGet(t *testing.T) {
	ds, cleanup := makeTempStore(t)
	defer cleanup()

	ctx := context.Background()
	doc, err := document.NewDocument("doc1", []embeddings.Chunk{{
		Embedding: []float32{1, 0},
		Block:     "empty",
	}}, "empty", time.Now(), "path/example")
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

	d1, _ := document.NewDocument("a", []embeddings.Chunk{
		{
			Embedding: []float32{0.5, 0.5},
			Block:     "block1",
		},
		{
			Embedding: []float32{1.0, 0.0},
			Block:     "block2",
		},
	}, "", time.Now(), "path/example")
	d2, _ := document.NewDocument("b", []embeddings.Chunk{
		{
			Embedding: []float32{1.0, 0.0},
			Block:     "block1",
		},
		{
			Embedding: []float32{0.1, 1.0},
			Block:     "block2",
		},
	}, "", time.Now(), "path/example")

	assert.NoError(t, ds.Index(ctx, d1))
	assert.NoError(t, ds.Index(ctx, d2))

	query := []float32{0.1, 0.9}

	results, err := ds.Search(ctx, query, 2)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	bestDoc := results[0]
	bestChunk := bestDoc.Document.Chunks[bestDoc.BestMatchingChunk]
	assert.Equal(t, "b", bestDoc.Document.ID)
	assert.Equal(t, "block2", bestChunk.Block)
	assert.True(t, vector.CosineSimilarity(query, bestChunk.Embedding) > results[1].Score)
}

func TestSearchEmpty(t *testing.T) {
	ds, cleanup := makeTempStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := ds.Search(ctx, []float32{1, 2, 3}, 1)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestList(t *testing.T) {
	ds, cleanup := makeTempStore(t)
	defer cleanup()

	ctx := context.Background()
	d1, _ := document.NewDocument("a", []embeddings.Chunk{
		{
			Embedding: []float32{0.5, 0.5},
			Block:     "block1",
		},
		{
			Embedding: []float32{1.0, 0.0},
			Block:     "block2",
		},
	}, "", time.Now(), "path/example")
	d2, _ := document.NewDocument("b", []embeddings.Chunk{
		{
			Embedding: []float32{1.0, 0.0},
			Block:     "block1",
		},
		{
			Embedding: []float32{0.1, 1.0},
			Block:     "block2",
		},
	}, "", time.Now(), "path/example")

	assert.NoError(t, ds.Index(ctx, d1))
	assert.NoError(t, ds.Index(ctx, d2))

	res, err := ds.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(res), 2)
	assert.Equal(t, res, []document.Document{d1, d2})
}

func TestPersistenceAcrossLoads(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "store.jsonl")

	{ // first load
		ds, err := NewDiskStore(file)
		assert.NoError(t, err)
		defer ds.Close()

		doc, _ := document.NewDocument("persist", []embeddings.Chunk{{
			Embedding: []float32{0.5, 0.5},
			Block:     "This is a test chunk",
		}}, "content", time.Now(), "path/example")
		assert.NoError(t, ds.Index(context.Background(), doc))
	}

	{ // second load
		ds2, err := NewDiskStore(file)
		assert.NoError(t, err)
		defer ds2.Close()

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

func TestRemove(t *testing.T) {
	doc1 := document.Document{ID: "id1", Content: "data1"}
	doc2 := document.Document{ID: "id2", Content: "data2"}
	doc3 := document.Document{ID: "id3", Content: "data3"}
	docs := createDocumentsOnDisk(t, []document.Document{doc1, doc2, doc3})

	dir := t.TempDir()
	file := filepath.Join(dir, "store.jsonl")

	{ // first load
		ds, err := NewDiskStore(file)
		assert.NoError(t, err)
		defer ds.Close()

		for _, doc := range docs {
			assert.NoError(t, ds.Index(context.Background(), doc))
		}
		assert.NoError(t, ds.Remove(context.Background(), "id2"))
	}

	{ // second load
		ds2, err := NewDiskStore(file)
		assert.NoError(t, err)
		defer ds2.Close()

		foundDocs, err := ds2.List(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 2, len(foundDocs))

		_, err = ds2.Get(context.Background(), "id2")
		assert.ErrorIs(t, ErrNotFound, err)
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

func createDocumentsOnDisk(t *testing.T, initialDocs []document.Document) []document.Document {
	t.Helper()

	tempDir := t.TempDir()

	docs := make([]document.Document, 0, len(initialDocs))
	for i, doc := range initialDocs {
		tempFilePath := filepath.Join(tempDir, fmt.Sprintf("test_data_%d.txt", i))
		doc.Path = tempFilePath
		docs = append(docs, doc)
	}

	return docs
}
