package cmd

import (
	"fmt"
	"time"

	"github.com/jnaraujo/seekr/internal/config"
	"github.com/jnaraujo/seekr/internal/storage"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display SeekR statistics.",
	Run: func(cmd *cobra.Command, args []string) {
		printAscii()

		fmt.Printf("Startup: %s\n", time.Since(startTime))
		fmt.Printf("Version: %g\n", config.AppVersion)

		storePath, err := storage.DefaultStorePath()
		if err == nil {
			fmt.Printf("Store Path: %s\n", storePath)
		}

		docs, _ := store.List(cmd.Context())
		fmt.Printf("Total Documents: %d\n", len(docs))

		fmt.Println("\nSeekR is running smoothly!")
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
