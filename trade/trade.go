package trade

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/modood/cts/dingtalk"
	"github.com/modood/cts/gateio"
	"github.com/modood/cts/util"
	"github.com/pkg/errors"
)

// Status
const (
	None  = -1 // 未知
	Empty = 0  // 空仓
	Full  = 1  // 满仓
)

var balance *gateio.Balance
var order *gateio.Order = &gateio.Order{}

// Flush refresh balance cache
func Flush(currency string) error {
	b, err := gateio.MyBalance()
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	balance = b

	// check latest deal
	o, err := gateio.LatestOrder(currency)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}
	if o.TradeID != "" && o.TradeID != order.TradeID {
		// have new deal
		text, err := message(o)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
		err = dingtalk.Push(text)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
		order = o
	}

	return nil
}

// Position return my open interest status
func Position(currency string) (int8, error) {
	usdt, err := carry("USDT")
	if err != nil {
		return None, errors.Wrap(err, util.FuncName())
	}

	amount, err := carry(currency)
	if err != nil {
		return None, errors.Wrap(err, util.FuncName())
	}

	pair, err := gateio.Ticker(currency)
	if err != nil {
		return None, errors.Wrap(err, util.FuncName())
	}

	if usdt < 10 {
		return Full, nil
	} else if amount*pair.Last < 10 {
		return Empty, nil
	}
	return None, nil
}

// AllIn all in
func AllIn(currency string) error {
	count, err := gateio.OpenOrderLen()
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}
	if count != 0 {
		err := gateio.Cancel(currency, gateio.CancelTypeAll)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
	}

	p, err := Position(currency)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	if p == Full {
		return nil
	}

	usdt, err := carry("USDT")
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	pair, err := gateio.Ticker(currency)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	amount := usdt / pair.HighestBid / 1.002 // fee: 0.2%
	if amount < 0.001 {
		return nil
	}

	log.Println("buying", currency, pair.HighestBid, amount)
	_, err = gateio.Buy(currency, pair.HighestBid, amount)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

// AllOut all out
func AllOut(currency string) error {
	count, err := gateio.OpenOrderLen()
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}
	if count != 0 {
		err := gateio.Cancel(currency, gateio.CancelTypeAll)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
	}

	p, err := Position(currency)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	if p == Empty {
		return nil
	}

	amount, err := carry(currency)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	pair, err := gateio.Ticker(currency)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	amount = amount / 1.002 // fee: 0.2%
	if amount < 0.001 {
		return nil
	}

	log.Println("selling", currency, pair.LowestAsk, amount)
	_, err = gateio.Sell(currency, pair.LowestAsk, amount)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

func carry(currency string) (float64, error) {
	coin := currency
	if coin != "USDT" {
		coin = strings.ToUpper(strings.Split(currency, "_")[0])
	}

	if balance == nil {
		if err := Flush(currency); err != nil {
			return 0, errors.Wrap(err, util.FuncName())
		}
	}

	str, ok := balance.Available[coin]
	if !ok {
		return 0, nil
	}

	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	return f, nil
}

func message(o *gateio.Order) (string, error) {
	a, err := gateio.MyAsset()
	if err != nil {
		return "", errors.Wrap(err, util.FuncName())
	}

	rise, fall, err := gateio.Trend()
	if err != nil {
		return "", errors.Wrap(err, util.FuncName())
	}

	return strconv.Itoa(time.Now().Year()) + "-" + o.Date +
		"\n行情：" + strconv.FormatUint(uint64(rise), 10) + "↑, " + strconv.FormatUint(uint64(fall), 10) + "↓" +
		"\n订单：" + o.TradeID +
		"\n类型：" + o.Type +
		"\n品种：" + o.Currency +
		"\n数量：" + o.Amount +
		"\n价格：$" + o.Rate +
		"\n金额：$" + strconv.FormatFloat(o.Total, 'f', 2, 64) +
		"\n挂单：$" + strconv.FormatFloat(a.Pending, 'f', 2, 64) +
		"\n余额：$" + strconv.FormatFloat(a.Balance, 'f', 2, 64) +
		"\n资金：$" + strconv.FormatFloat(a.Total, 'f', 2, 64) +
		"\n合计：¥" + strconv.FormatFloat(a.TotalCNY, 'f', 2, 64), nil
}
