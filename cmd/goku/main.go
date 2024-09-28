package main

import (
	"github.com/fallrising/goku-cli/cmd/goku/commands"
	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/fallrising/goku-cli/internal/database"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func init() {
	// Set up logging to a file
	logFile, err := os.OpenFile("goku.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	// Set log output to the file
	log.SetOutput(logFile)

	// Optionally, set the log flags for more detailed logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	dbPath := os.Getenv("GOKU_DB_PATH")
	if dbPath == "" {
		dbPath = "goku.db" // Default to current directory if not specified
	}

	cacheDBPath := os.Getenv("GOKU_CACHE_DB_PATH")
	if cacheDBPath == "" {
		cacheDBPath = "goku_cache.db" // Default cache database path
	}

	db, err := database.NewDatabase(dbPath, cacheDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	bookmarkService := bookmarks.NewBookmarkService(db)

	app := &cli.App{
		Name:  "goku",
		Usage: "A powerful CLI bookmark manager",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "db",
				EnvVars: []string{"GOKU_DB_PATH"},
				Value:   "goku.db",
				Usage:   "Path to the Goku database file",
			},
			&cli.StringFlag{
				Name:    "cache-db",
				EnvVars: []string{"GOKU_CACHE_DB_PATH"},
				Value:   "goku_cache.db",
				Usage:   "Path to the Goku cache database file",
			},
		},
		Commands: []*cli.Command{
			commands.AddCommand(bookmarkService),
			commands.DeleteCommand(bookmarkService),
			commands.GetCommand(bookmarkService),
			commands.ListCommand(bookmarkService),
			commands.SearchCommand(bookmarkService),
			commands.UpdateCommand(bookmarkService),
			commands.ImportCommand(bookmarkService),
			commands.ExportCommand(bookmarkService),
			commands.TagsCommand(bookmarkService),
			commands.StatsCommand(bookmarkService),
			commands.PurgeCommand(bookmarkService),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
