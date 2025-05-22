package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/jnaraujo/seekr/internal/id"
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

		if err := store.Remove(cmd.Context(), id.HashPath(inputPath)); err != nil {
			fmt.Printf("failed to remove document %q: %v\n", inputPath, err)
			return
		}

		fmt.Println("Document removed successfully")
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
