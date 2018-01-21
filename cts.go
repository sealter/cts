package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "cts"
	app.Usage = "the coin trading strategy"
	app.Version = "0.0.1"
	app.Author = "modood - https://github.com/modood"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "key",
			Usage: "your api key",
		},
		cli.StringFlag{
			Name:  "secret",
			Usage: "your secret key",
		},
	}
	app.Action = func(c *cli.Context) error {
		fmt.Println("doing")
		return nil
	}

	app.Run(os.Args)
}
