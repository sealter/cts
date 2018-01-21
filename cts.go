package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/modood/cts/gateio"
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
		for {
			// TODO
			btc, err := gateio.Ticker("btc_usdt")
			if err != nil {
				return err
			}

			log.Println(
				fmt.Sprintf("BTC/USDT $%f, 24h Change: %f%%",
					btc.Last, btc.PercentChange))

			time.Sleep(time.Second * 10)
		}
		return nil
	}

	app.Run(os.Args)
}
