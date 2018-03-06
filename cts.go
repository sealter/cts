package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/modood/cts/dingtalk"
	"github.com/modood/cts/gateio"
	"github.com/modood/cts/huobi"
	"github.com/modood/cts/strategy"
	"github.com/modood/cts/trade"
	"github.com/modood/cts/util"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/urfave/cli"
)

var strategies = strategy.Strategies()
var count uint64

func init() {
	tloc, err := time.LoadLocation("Asia/Chongqing")
	if err != nil {
		tloc = time.FixedZone("CST", 3600*8)
	}
	time.Local = tloc
}

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
			Name:  "gkey",
			Usage: "your gateio api key",
		},
		cli.StringFlag{
			Name:  "gsecret",
			Usage: "your gateio secret key",
		},
		cli.StringFlag{
			Name:  "hkey",
			Usage: "your huobi api key",
		},
		cli.StringFlag{
			Name:  "hsecret",
			Usage: "your huobi secret key",
		},
		cli.StringFlag{
			Name:  "dingtoken",
			Usage: "your access token of dingtalk group chat robot",
		},
	}
	app.Action = action
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

func action(c *cli.Context) error {
	log.Println("running...")

	gateio.Init(c.String("gkey"), c.String("gsecret"))
	huobi.Init(c.String("hkey"), c.String("hsecret"))
	dingtalk.Init(c.String("dingtoken"))
	currency := c.String("currency")
	stra := c.String("strategy")

	// cron job
	cr := schedule()
	cr.Start()

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
		return strategy.SigNone, errors.Wrap(err, util.FuncName())
	}

	sig, err := s.Signal()
	if err != nil {
		return strategy.SigNone, errors.Wrap(err, util.FuncName())
	}

	return sig, nil
}

func exec(signal uint8, currency string) error {
	s, err := huobi.NewSymbol(currency)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	switch signal {
	case strategy.SigRise:
		err := trade.AllIn(currency)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
		err = s.AllIn("BUY", false)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
	case strategy.SigFall:
		err := trade.AllOut(currency)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
		err = s.AllIn("SELL", false)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
	case strategy.SigBull:
		err := trade.AllIn(currency)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
		err = s.AllIn("BUY", true)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
	case strategy.SigBear:
		err := trade.AllOut(currency)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
		err = s.AllIn("SELL", true)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
	case strategy.SigNone:
		fallthrough
	default:
		// do nothing
	}
	return nil
}

func handle(err error) {
	atomic.AddUint64(&count, 1)
	log.Println(err)
}

func schedule() *cron.Cron {
	c := cron.New()
	err := c.AddFunc("0 0 7-23,0 * * *", func() {
		var retry uint16

	asset:
		a, err := gateio.MyAsset()
		if err != nil {
			if retry++; retry < 5 {
				goto asset
			}
			e := dingtalk.Push(err.Error())
			if e != nil {
				log.Println(e, err)
			}
		}

		retry = 0

	trend:
		rise, fall, err := gateio.Trend()
		if err != nil {
			if retry++; retry < 5 {
				goto trend
			}
			e := dingtalk.Push(err.Error())
			if e != nil {
				log.Println(e, err)
			}
		}

		msg := fmt.Sprintf("监控：%d Error(s)\n行情：%d↑, %d↓\n挂单：$%.2f\n余额：$%.2f\n资金：$%.2f\n合计：¥%.2f",
			atomic.LoadUint64(&count), rise, fall, a.Pending, a.Balance, a.Total, a.TotalCNY)
		atomic.StoreUint64(&count, 0)

		err = dingtalk.Push(msg)
		if err != nil {
			log.Println(err)
		}
	})
	if err != nil {
		e := dingtalk.Push(err.Error())
		if e != nil {
			log.Println(e, err)
		}
	}

	return c
}
