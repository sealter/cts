package trade

import (
	"errors"
	"strconv"

	"github.com/modood/cts/gateio"
)

const (
	None  = -1 // 未知
	Empty = 0  // 空仓
	Full  = 1  // 满仓
)

// Position return my open interest status
func Position() (int8, error) {
	b, err := gateio.MyBalance()
	if err != nil {
		return None, err
	}

	usdt, ok := b.Available["USDT"]
	if !ok {
		return None, errors.New("Have no USDT")
	}

	f, err := strconv.ParseFloat(usdt, 64)
	if err != nil {
		return None, err
	}

	if f > 10 {
		return Empty, nil
	}
	return Full, nil
}
