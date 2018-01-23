package main

import (
	"log"
	"os"
	"time"

	"github.com/modood/cts/gateio"
	"github.com/modood/cts/strategy"
	"github.com/modood/cts/trade"
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
		gateio.Init("apikey", "secretkey")
		for {
			time.Sleep(time.Second * 1)

			err := trade.Flush()
			if err != nil {
				log.Println(err)
				continue
			}

			sig, err := strategy.RippleDoge()
			if err != nil {
				log.Println(err)
				continue
			}

			switch sig {
			case strategy.SIG_RISE:
				err := trade.AllIn("btc_usdt")
				if err != nil {
					log.Println(err)
					break
				}
			case strategy.SIG_FALL:
				err := trade.AllOut("btc_usdt", "BTC")
				if err != nil {
					log.Println(err)
					break
				}
			case strategy.SIG_NONE:
				fallthrough
			default:
				log.Println("do nothing")
			}
		}
		return nil
	}

	app.Run(os.Args)
}
