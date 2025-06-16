package commands

import (
	"context"
	"fmt"
	"github.com/fallrising/goku-cli/internal/bookmarks"
	"github.com/fallrising/goku-cli/internal/mqtt"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

func ImportCommand() *cli.Command {
	return &cli.Command{
		Name: "import",
		Usage: "Import bookmarks from HTML, JSON, or plain text URL list\n\n" +
			"Examples:\n" +
			"  goku import --file bookmarks.html\n" +
			"  goku import -f bookmarks.json --workers 10\n" +
			"  goku import --file bookmarks.txt\n" +
			"  goku import -f urls.txt --mqtt-broker localhost --mqtt-port 1883 --mqtt-topic bookmarks/imported",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Input file path (.html, .json, or .txt)",
				Required: true,
			},
			&cli.IntFlag{
				Name:    "workers",
				Aliases: []string{"w"},
				Usage:   "Number of worker goroutines for concurrent processing",
				Value:   5, // Default value
			},
			&cli.BoolFlag{
				Name:    "fetch",
				Aliases: []string{"F"},
				Usage:   "Enable fetching additional data for each bookmark",
				Value:   false, // Disabled by default
			},
			// MQTT Configuration Flags
			&cli.StringFlag{
				Name:  "mqtt-broker",
				Usage: "MQTT broker hostname/IP (enables MQTT publishing)",
			},
			&cli.IntFlag{
				Name:  "mqtt-port",
				Usage: "MQTT broker port",
				Value: 1883,
			},
			&cli.StringFlag{
				Name:  "mqtt-client-id",
				Usage: "MQTT client ID (auto-generated if not provided)",
			},
			&cli.StringFlag{
				Name:  "mqtt-username",
				Usage: "MQTT username (optional)",
			},
			&cli.StringFlag{
				Name:  "mqtt-password",
				Usage: "MQTT password (optional)",
			},
			&cli.StringFlag{
				Name:  "mqtt-topic",
				Usage: "MQTT topic for bookmark events",
				Value: "goku/bookmarks",
			},
			&cli.IntFlag{
				Name:  "mqtt-qos",
				Usage: "MQTT QoS level (0, 1, or 2)",
				Value: 1,
			},
		},
		Action: func(c *cli.Context) error {
			filePath := c.String("file")
			numWorkers := c.Int("workers")
			fetchData := c.Bool("fetch")
			bookmarkService := c.App.Metadata["bookmarkService"].(*bookmarks.BookmarkService)

			// Setup MQTT client if broker is provided
			var mqttClient *mqtt.Client
			if mqttBroker := c.String("mqtt-broker"); mqttBroker != "" {
				mqttConfig := &mqtt.Config{
					Broker:   mqttBroker,
					Port:     c.Int("mqtt-port"),
					ClientID: c.String("mqtt-client-id"),
					Username: c.String("mqtt-username"),
					Password: c.String("mqtt-password"),
					Topic:    c.String("mqtt-topic"),
					QoS:      byte(c.Int("mqtt-qos")),
				}
				
				var err error
				mqttClient, err = mqtt.NewClient(mqttConfig)
				if err != nil {
					return fmt.Errorf("failed to create MQTT client: %w", err)
				}
				
				if err := mqttClient.Connect(); err != nil {
					return fmt.Errorf("failed to connect to MQTT broker: %w", err)
				}
				defer mqttClient.Disconnect()
				
				fmt.Printf("MQTT: Connected to broker %s:%d, publishing to topic '%s'\n", 
					mqttBroker, c.Int("mqtt-port"), c.String("mqtt-topic"))
			}

			// Open the file
			file, err := openFile(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			// Create a context with the import options
			ctx := context.WithValue(context.Background(), "numWorkers", numWorkers)
			ctx = context.WithValue(ctx, "fetchData", fetchData)
			ctx = context.WithValue(ctx, "mqttClient", mqttClient)

			// Determine import type based on file extension
			var recordsCreated int
			if isJSON(filePath) {
				recordsCreated, err = bookmarkService.ImportFromJSON(ctx, file)
			} else if isHTML(filePath) {
				recordsCreated, err = bookmarkService.ImportFromHTML(ctx, file)
			} else if isText(filePath) {
				recordsCreated, err = bookmarkService.ImportFromText(ctx, file)
			} else {
				return fmt.Errorf("unsupported file format: %s", filePath)
			}

			if err != nil {
				return fmt.Errorf("failed to import bookmarks: %w", err)
			}

			fmt.Printf("Import completed. %d bookmarks were successfully imported.\n", recordsCreated)
			if fetchData {
				fmt.Println("Additional data was fetched for each bookmark.")
			}
			if mqttClient != nil {
				fmt.Printf("MQTT: Published %d bookmark events to topic '%s'\n", recordsCreated, c.String("mqtt-topic"))
			}
			return nil
		},
	}
}

// openFile opens the file and returns an error if it fails.
func openFile(filePath string) (*os.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// isJSON checks if the file is a JSON file based on the file extension.
func isJSON(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".json")
}

// isHTML checks if the file is an HTML file based on the file extension.
func isHTML(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".html") || strings.HasSuffix(strings.ToLower(filePath), ".htm")
}

// isText checks if the file is a plain text file based on the file extension.
func isText(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".txt")
}
