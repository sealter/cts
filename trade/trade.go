package trade

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/modood/cts/gateio"
)

const (
	None  = -1 // 未知
	Empty = 0  // 空仓
	Full  = 1  // 满仓
)

var balance *gateio.Balance

// Flush refresh balance cache
func Flush() error {
	b, err := gateio.MyBalance()
	if err != nil {
		return err
	}

	balance = b

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

	amount := (usdt - 1) / pair.Last

	log.Println("buying", currency, pair.Last, amount)
	// _, err := Buy(currency, pair.Last, amount)

	return err
}

// AllOut all out
func AllOut(currency, coin string) error {
	p, err := Position()
	if err != nil {
		return err
	}
	if p == Empty {
		return nil
	}

	amount, err := carry(coin)
	if err != nil {
		return err
	}

	pair, err := gateio.Ticker(currency)
	if err != nil {
		return err
	}

	log.Println("selling", currency, pair.Last, amount)
	// _, err := Sell(currency, pair.Last, amount)

	return err
}

func carry(coin string) (float64, error) {
	if balance == nil {
		err := Flush()

		if err != nil {
			return None, err
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
