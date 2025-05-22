package cmd

import (
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
		if err := storage.CheckFile(inputPath); err != nil {
			fmt.Printf("failed to index document %q: %v\n", inputPath, err)
			return
		}

		inputPath, err := filepath.Abs(inputPath)
		if err != nil {
			fmt.Printf("failed to get absolute path for %q: %v\n", inputPath, err)
			return
		}

		contentBytes, err := os.ReadFile(inputPath)
		if err != nil {
			fmt.Printf("failed to read document %q: %v\n", inputPath, err)
			return
		}

		content := string(contentBytes)
		if len(content) == 0 {
			fmt.Printf("document %q is empty\n", inputPath)
			return
		}
		if len(content) > config.MaxContent {
			fmt.Printf("document %q is too large (%d bytes)\n", inputPath, len(content))
			return
		}

		if store.HasPath(cmd.Context(), inputPath) {
			fmt.Printf("document %q is already indexed\n", inputPath)
			return
		}

		fmt.Printf("Indexing document %q...\n", inputPath)
		chunks, err := embedding.Embed(cmd.Context(), content)
		if err != nil {
			fmt.Printf("failed to generate document embeddings %q: %v\n", inputPath, err)
			return
		}

		doc, err := document.NewDocument(id.NewID(), chunks, content, time.Now(), inputPath)
		if err != nil {
			fmt.Printf("failed to create document %q: %v\n", inputPath, err)
			return
		}

		err = store.Index(cmd.Context(), doc)
		if err != nil {
			fmt.Printf("failed to index document %q: %v\n", inputPath, err)
			return
		}

		fmt.Printf("Document %q indexed successfully!\n", inputPath)
		fmt.Printf("Document ID: %s\n", doc.ID)
		fmt.Printf("Document Size: %d bytes\n", len(contentBytes))
		fmt.Printf("Document Path: %s\n", doc.Path)
		fmt.Printf("Document Created At: %s\n", doc.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Document Chunk Size: %d\n", len(doc.Chunks))
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
