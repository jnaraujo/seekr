package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

		switch pathKind {
		case storage.DirectoryPathKind:
			fmt.Printf("Indexing directory %q...\n", inputPath)
			files, err := storage.FilePathWalkDir(inputPath)
			if err != nil {
				fmt.Printf("Failed to index directory %q: %v\n", inputPath, err)
				return
			}

			for _, file := range files {
				fmt.Printf("Indexing document %q...\n", file)
				err = indexFile(cmd.Context(), file)
				if err != nil {
					fmt.Printf("Failed to index %q: %v\n", file, err)
					continue
				}
				fmt.Printf("Document %q indexed successfully!\n", file)
			}
			fmt.Printf("Directory %q indexed.\n", inputPath)
		case storage.FilePathKind:
			fmt.Printf("Indexing document %q...\n", inputPath)
			err = indexFile(cmd.Context(), inputPath)
			if err != nil {
				fmt.Printf("Failed to index %q: %v\n", inputPath, err)
				return
			}
			fmt.Printf("Document %q indexed successfully!\n", inputPath)
		default:
			fmt.Printf("Failed to index document %q\n", inputPath)
			return
		}

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
