package trade

import (
	"errors"
	"log"
	"strconv"

	"github.com/modood/cts/gateio"
)

const (
	None  = -1 // 未知
	Empty = 0  // 空仓
	Full  = 1  // 满仓
)

// Position return my open interest status
func Position() (position int8, balance float64, e error) {
	b, err := gateio.MyBalance()
	if err != nil {
		return None, 0, err
	}

	usdt, ok := b.Available["USDT"]
	if !ok {
		return None, 0, errors.New("Have no USDT")
	}

	f, err := strconv.ParseFloat(usdt, 64)
	if err != nil {
		return None, 0, err
	}

	if f > 10 {
		return Empty, f, nil
	}
	return Full, f, nil
}

// Allin all in
func AllIn(currency string) error {
	p, f, err := Position()
	if err != nil {
		return err
	}

	if p == Full {
		return nil
	}

	pair, err := gateio.Ticker(currency)
	if err != nil {
		return err
	}

	amount := (f - 1) / pair.Last

	log.Println("buying", currency, pair.Last, amount)
	// _, err := Buy(currency, pair.Last, amount)

	return err
}
