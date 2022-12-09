package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "tacher",
		Version:     "v0.1",
		HideVersion: false,
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "init a Spring Boot project",
				Action: func(ctx *cli.Context) error {
					err := RunUI(
						ctx.String("group"),
						ctx.String("artifact"),
						ctx.String("name"),
						ctx.String("description"),
						ctx.String("package"),
					)
					if err != nil {
						return fmt.Errorf("An error occured while setting up the application: %w", err)
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "group",
						Usage:    "Group",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "artifact",
						Usage:    "Artifact",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Name",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "description",
						Usage:    "Description",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "package",
						Usage:    "Package name",
						Required: false,
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
