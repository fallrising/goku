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
		dbPath = "goku.db"
	}

	cacheDBPath := os.Getenv("GOKU_CACHE_DB_PATH")
	if cacheDBPath == "" {
		cacheDBPath = "goku_cache.db"
	}

	duckDBPath := os.Getenv("GOKU_DUCKDB_PATH")
	if duckDBPath == "" {
		duckDBPath = "goku_stats.duckdb"
	}

	db, err := database.NewDatabase(dbPath, cacheDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	duckDBStats, err := database.NewDuckDBStats(duckDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize DuckDB: %v", err)
	}

	if err := duckDBStats.Init(); err != nil {
		log.Fatalf("Failed to initialize DuckDB schema: %v", err)
	}

	bookmarkService := bookmarks.NewBookmarkService(db, duckDBStats)

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
			&cli.StringFlag{
				Name:    "duckdb",
				EnvVars: []string{"GOKU_DUCKDB_PATH"},
				Value:   "goku_stats.duckdb",
				Usage:   "Path to the Goku DuckDB statistics file",
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
			commands.SyncCommand(bookmarkService), // Add a new command to sync data to DuckDB
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
