package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all available documents stored in the system.",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		docs, err := store.List(cmd.Context())
		if err != nil {
			fmt.Printf("Error listing documents: %v\n", err)
			return
		}

		if len(docs) == 0 {
			fmt.Println("No documents found.")
			return
		}
		fmt.Printf("Found %d document(s)\n\n", len(docs))
		fmt.Println("# - Content - Path")
		fmt.Println("-------------------")

		maxDigits := countDigits(len(docs))
		for index, doc := range docs {
			title := ""
			if len(doc.Content) > 30 {
				title = doc.Content[:30] + "..."
			} else {
				title = doc.Content
			}
			fmt.Printf("%0*d - %s - %s\n", index, maxDigits, title, doc.Path)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func countDigits(number int) int {
	return len(strconv.Itoa(number))
}
