package main

import (
	"log"
	"os"

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
			// ... (rest of the CLI commands remain the same)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
