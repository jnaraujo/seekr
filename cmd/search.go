package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search",
	Short:   "Search documents",
	Long:    "Search for documents based on a query string.",
	Example: "seekr search 'your query here'",
	Aliases: []string{"find"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		chunks, err := embedding.Embed(cmd.Context(), query)
		if err != nil {
			fmt.Printf("failed to create embedding: %v\n", err)
			return
		}
		results, err := store.Search(cmd.Context(), chunks[0].Embedding, 5)
		if err != nil {
			fmt.Printf("failed to search documents: %v\n", err)
			return
		}

		if len(results) == 0 {
			fmt.Println("No results found.")
			return
		}

		fmt.Println("(#) %% sim - Content - Path")
		fmt.Println("-----------------------------")
		for index, res := range results {
			title := ""
			if len(res.Document.Content) > 30 {
				title = res.Document.Content[:30] + "..."
			} else {
				title = res.Document.Content
			}
			fmt.Printf("(%d) %.2f%% - %s - %s\n", index+1, res.Score*100, title, res.Document.Path)
		}
		fmt.Printf("\nFound %d results.\n", len(results))
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
