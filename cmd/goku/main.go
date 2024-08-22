package main

import (
	"fmt"
	"os"

	"github.com/fallrising/goku/internal/database"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "goku",
		Short: "A command-line bookmark manager.",
		Long:  `Goku is a fast and flexible tool for managing your bookmarks from the command line.`,
	}

	// Initialize the database
	var err error
	database.Db, err = database.InitDB() // Update to database.InitDB()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		os.Exit(1)
	}
	defer database.Db.Close()

	// Add subcommands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
