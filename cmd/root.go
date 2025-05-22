package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jnaraujo/seekr/internal/config"
	"github.com/jnaraujo/seekr/internal/storage"
	"github.com/spf13/cobra"
)

var startTime = time.Now()
var store storage.Store

var rootCmd = &cobra.Command{
	Use:   "seekr",
	Short: "SeekR is a lightweight yet powerful semantic search engine for local data.",
	Long: `A simple yet powerful local semantic search engine that brings intelligent, context-aware search capabilities to your data—using locally generated embeddings via Ollama or any compatible embedding provider, with no reliance on external APIs.

Learn more at https://github.com/jnaraujo/seekr
`,
	Run: func(cmd *cobra.Command, args []string) {
		printAscii()
		fmt.Println("Welcome to SeekR!")
		fmt.Println("Use --help to see available commands.")
	},
}

func Execute() {
	storePath, err := storage.DefaultStorePath()
	if err != nil {
		fmt.Println("Error getting default store path:", err)
		return
	}
	if err := storage.EnsureStorePath(storePath); err != nil {
		fmt.Println("Error ensuring store path:", err)
		return
	}

	store, err = storage.NewDiskStore(storePath)
	if err != nil {
		fmt.Println("Error creating disk store:", err)
		return
	}
	defer store.Close()

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printAscii() {
	art := `
  _________              __   __________ 
 /   _____/ ____   ____ |  | _\______   \
 \_____  \_/ __ \_/ __ \|  |/ /|       _/
 /        \  ___/\  ___/|    < |    |   \
/_______  /\___  >\___  >__|_ \|____|_  /
        \/     \/     \/     \/       \/ 
`
	fmt.Println(art)
	bottom := fmt.Sprintf("%s - v%g", config.AppName, config.AppVersion)
	const asciiWidth = 41
	if len(bottom) < asciiWidth {
		padding := (asciiWidth - len(bottom)) / 2
		fmt.Printf("%s%s\n\n", strings.Repeat(" ", padding), bottom)
	} else {
		fmt.Printf("%s\n\n", bottom)
	}
}
