package main

import (
	"log"
	"os"
	"sort"

	"github.com/fallrising/goku-cli/cmd/goku/commands"
	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/fallrising/goku-cli/internal/database"
	"github.com/urfave/cli/v2"
)

func init() {
	setupLogging()
}

func main() {
	app := createApp()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func setupLogging() {
	logFile, err := os.OpenFile("goku.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func createApp() *cli.App {
	bookmarkService := setupDatabases()

	app := &cli.App{
		Name:        "goku",
		Usage:       "A powerful CLI bookmark manager",
		Description: "Goku CLI helps you manage your bookmarks efficiently from the command line.",
		Version:     "1.0.0",
		Authors: []*cli.Author{
			{
				Name:  "KC",
				Email: "",
			},
		},
		Flags:    getGlobalFlags(),
		Commands: getCommands(bookmarkService),
	}

	sort.Sort(cli.CommandsByName(app.Commands))
	cli.AppHelpTemplate = getCustomAppHelpTemplate()

	return app
}

func setupDatabases() *bookmarks.BookmarkService {
	dbPath := getEnvOrDefault("GOKU_DB_PATH", "goku.db")
	cacheDBPath := getEnvOrDefault("GOKU_CACHE_DB_PATH", "goku_cache.db")
	duckDBPath := getEnvOrDefault("GOKU_DUCKDB_PATH", "goku_stats.duckdb")

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

	return bookmarks.NewBookmarkService(db, duckDBStats)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getGlobalFlags() []cli.Flag {
	return []cli.Flag{
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
	}
}

func getCommands(bookmarkService *bookmarks.BookmarkService) []*cli.Command {
	return []*cli.Command{
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
		commands.SyncCommand(bookmarkService),
		commands.FetchCommand(bookmarkService),
	}
}

func getCustomAppHelpTemplate() string {
	return `NAME:
   {{.Name}} - {{.Usage}}

DESCRIPTION:
   {{.Description}}

USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}

VERSION:
   {{.Version}}

AUTHOR:
   {{range .Authors}}{{ . }}{{end}}

COMMANDS:
{{range .Commands}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}
{{end}}
GLOBAL OPTIONS:
{{range .VisibleFlags}}   {{.}}
{{end}}`
}
