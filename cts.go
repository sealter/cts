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
	"github.com/modood/cts/util"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/urfave/cli"
)

var (
	strategies = strategy.Strategies()
	count      uint64
)

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
			Name:  "symbol",
			Usage: "symbol name, quote symbol should be usdt(e.g., btc_usdt, xrp_usdt, etc.)",
		},
		cli.StringFlag{
			Name:  "strategy",
			Usage: "strategy name. available: " + strings.Join(strategy.Available(), ", "),
		},
		cli.StringFlag{
			Name:  "key",
			Usage: "your huobi api key",
		},
		cli.StringFlag{
			Name:  "secret",
			Usage: "your huobi api secret",
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

	huobi.Init(c.String("key"), c.String("secret"))
	dingtalk.Init(c.String("dingtoken"))
	symbol := c.String("symbol")
	stra := c.String("strategy")

	// cron job
	cr := schedule()
	cr.Start()

	for {
		time.Sleep(time.Second * 5)

		sig, err := signal(stra)
		if err != nil {
			handle(err)
			continue
		}

		err = exec(sig, symbol)
		if err != nil {
			handle(err)
			continue
		}
	}
}

func signal(str string) (uint8, error) {
	s, ok := strategies[str]
	if !ok {
		err := fmt.Errorf("unknown strategy: %s", str)
		return strategy.SigNone, errors.Wrap(err, util.FuncName())
	}

	sig, err := s.Signal()
	if err != nil {
		return strategy.SigNone, errors.Wrap(err, util.FuncName())
	}

	return sig, nil
}

func exec(signal uint8, symbol string) error {
	s, err := huobi.NewSymbol(symbol)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	switch signal {
	case strategy.SigRise:
		err = s.AllIn("BUY", false)
	case strategy.SigFall:
		err = s.AllIn("SELL", false)
	case strategy.SigBull:
		err = s.AllIn("BUY", true)
	case strategy.SigBear:
		err = s.AllIn("SELL", true)
	case strategy.SigNone:
		fallthrough
	default:
		// do nothing
	}
	if err != nil {
		return errors.Wrap(err, util.FuncName())
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
		rise, fall, err := gateio.Trend()
		if err != nil {
			e := dingtalk.Push(err.Error())
			if e != nil {
				log.Println(e, err)
			}
			return
		}

		msg := fmt.Sprintf("%s\n监控：%d Error(s)\n行情：%d↑, %d↓",
			time.Now().Format("2006-01-02 15:04:05"),
			atomic.LoadUint64(&count), rise, fall)
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
