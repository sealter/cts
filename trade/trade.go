package trade

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/modood/cts/dingtalk"
	"github.com/modood/cts/gateio"
)

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
		return err
	}

	balance = b

	// check latest deal
	o, err := gateio.LatestOrder(currency)
	if err != nil {
		// do nothing
	} else if o.OrderID != "" && o.OrderID != order.OrderID {
		if order.OrderID != "" {
			// have new deal
			text, err := message(o)
			if err != nil {
				// do nothing
			}
			err = dingtalk.Push(text)
			if err != nil {
				// do nothing
			}
		}
		order = o
	}

	return nil
}

// Position return my open interest status
func Position() (int8, error) {
	usdt, err := carry("USDT")
	if err != nil {
		return None, err
	}

	if usdt > 10 {
		return Empty, nil
	}
	return Full, nil
}

// Allin all in
func AllIn(currency string) error {
	p, err := Position()
	if err != nil {
		return err
	}
	if p == Full {
		return nil
	}

	usdt, err := carry("USDT")
	if err != nil {
		return err
	}

	pair, err := gateio.Ticker(currency)
	if err != nil {
		return err
	}

	err = gateio.Cancel(currency, gateio.CancelTypeAll)
	if err != nil {
		return err
	}

	amount := (usdt - 1) / pair.Last
	_ = usdt

	log.Println("buying", currency, pair.Last, amount)
	_, err = gateio.Buy(currency, pair.Last, amount)

	return err
}

// AllOut all out
func AllOut(currency string) error {
	p, err := Position()
	if err != nil {
		return err
	}
	if p == Empty {
		return nil
	}

	amount, err := carry(currency)
	if err != nil {
		return err
	}

	pair, err := gateio.Ticker(currency)
	if err != nil {
		return err
	}

	err = gateio.Cancel(currency, gateio.CancelTypeAll)
	if err != nil {
		return err
	}

	log.Println("selling", currency, pair.Last, amount)
	_, err = gateio.Sell(currency, pair.Last, amount)

	return err
}

func carry(currency string) (float64, error) {
	coin := currency
	if coin != "USDT" {
		coin = strings.ToUpper(strings.Split(currency, "_")[0])
	}

	if balance == nil {
		if err := Flush(currency); err != nil {
			return 0, err
		}
	}

	str, ok := balance.Available[coin]
	if !ok {
		return 0, errors.New(fmt.Sprintf("have no %s", coin))
	}

	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, err
	}

	return f, nil
}

func message(o *gateio.Order) (string, error) {
	a, err := gateio.MyAsset()
	if err != nil {
		return "", err
	}

	return strconv.Itoa(time.Now().Year()) + "-" + o.Date +
		"\n订单：" + o.OrderID +
		"\n类型：" + o.Type +
		"\n品种：" + o.Currency +
		"\n数量：" + o.Amount +
		"\n价格：$" + o.Rate +
		"\n金额：$" + strconv.FormatFloat(o.Total, 'f', 2, 64) +
		"\n余额：$" + strconv.FormatFloat(a.Balance, 'f', 2, 64) +
		"\n资金：$" + strconv.FormatFloat(a.Total, 'f', 2, 64) +
		"\n合计：¥" + strconv.FormatFloat(a.TotalCNY, 'f', 2, 64), nil
}
