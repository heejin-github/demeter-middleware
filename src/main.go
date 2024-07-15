package main

import (
	"log"
	"os"
	"fmt"

	"github.com/urfave/cli/v2"
)

func main() {
	// Set up logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	app := &cli.App{
		Name:  "API Key Middleware",
		Usage: "Manage API keys and run middleware",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start the middleware",
				Action: func(c *cli.Context) error {
					log.Println("Starting middleware...")
					return startMiddleware()
				},
			},
			{
				Name:  "key",
				Usage: "Manage API keys",
				Subcommands: []*cli.Command{
					{
						Name:  "generate",
						Usage: "Generate a new API key",
						Action: func(c *cli.Context) error {
							apiKey := generateAPIKey()
							err := storeAPIKey(apiKey)
							if err != nil {
								return err
							}
							fmt.Printf("Generated API key: %s\n", apiKey)
							return nil
						},
					},
					{
						Name:  "list",
						Usage: "List all API keys",
						Action: func(c *cli.Context) error {
							keys, err := listAPIKeys()
							if err != nil {
								return err
							}
							for _, key := range keys {
								fmt.Println(key)
							}
							return nil
						},
					},
					{
						Name:  "delete",
						Usage: "Delete an API key",
						Action: func(c *cli.Context) error {
							if c.NArg() == 0 {
								return fmt.Errorf("Please provide an API key to delete")
							}
							apiKey := c.Args().First()
							err := deleteAPIKey(apiKey)
							if err != nil {
								return err
							}
							fmt.Printf("Deleted API key: %s\n", apiKey)
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
