package cmd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/jnaraujo/seekr/internal/config"
	"github.com/jnaraujo/seekr/internal/document"
	"github.com/jnaraujo/seekr/internal/id"
	"github.com/jnaraujo/seekr/internal/storage"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index the document",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputPath := args[0]
		pathKind, err := storage.CheckPath(inputPath)
		if err != nil {
			fmt.Printf("failed to index document %q: %v\n", inputPath, err)
			return
		}

		inputPath, err = filepath.Abs(inputPath)
		if err != nil {
			fmt.Printf("failed to get absolute path for %q: %v\n", inputPath, err)
			return
		}

		// parallelism for embeddings is not yet supported by Ollama, so I think this should be sufficient for now
		workers := int(math.Max(2, float64(runtime.NumCPU())))
		channel := make(chan string, workers)
		var wg sync.WaitGroup

		for range workers {
			go func() {
				for file := range channel {
					err := indexFile(cmd.Context(), file)
					if err != nil {
						fmt.Printf("Failed to index %q: %v\n", file, err)
					} else {
						fmt.Printf("Document %q indexed successfully!\n", file)
					}
					wg.Done()
				}
			}()
		}

		switch pathKind {
		case storage.DirectoryPathKind:
			fmt.Printf("Indexing directory %q...\n", inputPath)
			files, err := storage.FilePathWalkDir(inputPath)
			if err != nil {
				fmt.Printf("Failed to index directory %q: %v\n", inputPath, err)
			} else {
				for _, file := range files {
					wg.Add(1)
					channel <- file
				}
			}
		case storage.FilePathKind:
			wg.Add(1)
			channel <- inputPath
		default:
			fmt.Printf("Failed to index document %q\n", inputPath)
		}

		close(channel)
		wg.Wait()
		fmt.Println("Indexing complete!")
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
}

func indexFile(ctx context.Context, path string) error {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return errors.New("failed to read document")
	}

	// TODO: add PDF support
	if !storage.IsFileValid(contentBytes) {
		return fmt.Errorf("document is not a valid file type")
	}

	if storage.IsHidden(path) {
		return fmt.Errorf("hidden files are not supported")
	}

	content := string(contentBytes)
	if len(content) == 0 {
		return errors.New("document is empty")
	}
	if len(content) > config.MaxContent {
		return errors.New("document is too large")
	}

	if _, err := store.Get(ctx, id.HashPath(path)); err == nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return errors.New("document is already indexed")
		}
	}

	chunks, err := embedding.Embed(ctx, content)
	if err != nil {
		return errors.New("failed to generate document embeddings")
	}

	doc, err := document.NewDocument(id.HashPath(path), chunks, content, time.Now(), path)
	if err != nil {
		return errors.New("failed to create document ")
	}

	err = store.Index(ctx, doc)
	if err != nil {
		return errors.New("failed to index document")
	}

	return nil
}
