package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/modood/cts/gateio"
	"github.com/modood/cts/strategy"
	"github.com/modood/cts/trade"
	"github.com/urfave/cli"
)

var strategies = strategy.Strategies()

func main() {
	app := cli.NewApp()
	app.Name = "cts"
	app.Usage = "the coin trading strategy"
	app.Version = "0.0.1"
	app.Author = "modood - https://github.com/modood"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "currency",
			Usage: "currency name, base currency should be usdt(e.g., btc_usdt, xrp_usdt, etc.)",
		},
		cli.StringFlag{
			Name:  "strategy",
			Usage: "strategy name. available: " + strings.Join(strategy.Available(), ", "),
		},
		cli.StringFlag{
			Name:  "key",
			Usage: "your api key",
		},
		cli.StringFlag{
			Name:  "secret",
			Usage: "your secret key",
		},
	}
	app.Action = action
	app.Run(os.Args)
}

func action(c *cli.Context) error {
	log.Println("starting...")
	gateio.Init(c.String("key"), c.String("secret"))

	for {
		time.Sleep(time.Second * 1)

		err := trade.Flush()
		if err != nil {
			log.Println(err)
			continue
		}

		sig, err := signal(c.String("strategy"))
		if err != nil {
			log.Println(err)
			continue
		}

		err = exec(sig, c.String("currency"))
		if err != nil {
			log.Println(err)
			continue
		}
	}
	return nil
}

func signal(stra string) (uint8, error) {
	s, ok := strategies[stra]
	if !ok {
		return strategy.SIG_NONE,
			errors.New(fmt.Sprintf("unknown strategy: %s", stra))
	}

	sig, err := s.Signal()
	if err != nil {
		return strategy.SIG_NONE, err
	}

	return sig, nil
}

func exec(signal uint8, currency string) error {
	switch signal {
	case strategy.SIG_RISE:
		err := trade.AllIn(currency)
		if err != nil {
			return err
		}
	case strategy.SIG_FALL:
		coin := strings.ToUpper(strings.Split(currency, "_")[0])
		err := trade.AllOut(currency, coin)
		if err != nil {
			return err
		}
	case strategy.SIG_NONE:
		fallthrough
	default:
		log.Println("do nothing")
	}
	return nil
}
