package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/modood/cts/dingtalk"
	"github.com/modood/cts/gateio"
	"github.com/modood/cts/strategy"
	"github.com/modood/cts/trade"
	"github.com/modood/cts/util"
	"github.com/pkg/errors"
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
		cli.StringFlag{
			Name:  "dingtoken",
			Usage: "your access token of dingtalk group chat robot",
		},
	}
	app.Action = action
	app.Run(os.Args)
}

func action(c *cli.Context) error {
	log.Println("running...")

	gateio.Init(c.String("key"), c.String("secret"))
	dingtalk.Init(c.String("dingtoken"))
	currency := c.String("currency")
	stra := c.String("strategy")

	for {
		time.Sleep(time.Second * 5)

		err := trade.Flush(currency)
		if err != nil {
			handle(err)
			continue
		}

		sig, err := signal(stra)
		if err != nil {
			handle(err)
			continue
		}

		err = exec(sig, currency)
		if err != nil {
			handle(err)
			continue
		}
	}
}

func signal(stra string) (uint8, error) {
	s, ok := strategies[stra]
	if !ok {
		err := fmt.Errorf("unknown strategy: %s", stra)
		return strategy.SIG_NONE, errors.Wrap(err, util.FuncName())
	}

	sig, err := s.Signal()
	if err != nil {
		return strategy.SIG_NONE, errors.Wrap(err, util.FuncName())
	}

	return sig, nil
}

func exec(signal uint8, currency string) error {
	switch signal {
	case strategy.SIG_RISE:
		err := trade.AllIn(currency)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
	case strategy.SIG_FALL:
		err := trade.AllOut(currency)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
	case strategy.SIG_NONE:
		fallthrough
	default:
		// do nothing
	}
	return nil
}

func handle(err error) {
	e := dingtalk.Push(err.Error())
	if e != nil {
		log.Println(e, err)
	}
}
