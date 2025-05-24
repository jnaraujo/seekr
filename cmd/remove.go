package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/jnaraujo/seekr/internal/id"
	"github.com/jnaraujo/seekr/internal/storage"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Removes the document",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputPath := args[0]
		inputPath, err := filepath.Abs(inputPath)
		if err != nil {
			fmt.Printf("failed to get absolute path for %q: %v\n", inputPath, err)
			return
		}

		pathKind, _ := storage.CheckPath(inputPath)
		// we ignore the error because the document may not necessarily exist anymore, but we might still want to remove it from the index.
		switch pathKind {
		case storage.DirectoryPathKind:
			fmt.Printf("Removing directory %q...\n", inputPath)

			files, err := storage.FilePathWalkDir(inputPath)
			if err != nil {
				fmt.Printf("Failed to remove directory %q: %v\n", inputPath, err)
				return
			}

			for _, file := range files {
				fmt.Printf("Removing document %q...\n", file)
				err = store.Remove(cmd.Context(), id.HashPath(file))
				if err != nil {
					fmt.Printf("Failed to remove %q: %v\n", file, err)
					continue
				}
				fmt.Printf("Document %q removed successfully!\n", file)
			}

			fmt.Printf("Directory %q removed\n", inputPath)
		default:
			fmt.Printf("Removing file %q...\n", inputPath)
			err := store.Remove(cmd.Context(), id.HashPath(inputPath))
			if err != nil {
				fmt.Printf("failed to remove document %q: %v\n", inputPath, err)
				return
			}
			fmt.Println("Document removed successfully")
		}

	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
