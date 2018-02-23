package strategy

import (
	"log"
	"strconv"
	"strings"

	"github.com/modood/cts/gateio"
	"github.com/modood/cts/util"
	"github.com/pkg/errors"
)

// RippleDoge is a simple strategy refers to ripple and doge
type RippleDoge struct{}

// Name return strategy name
func (s RippleDoge) Name() string {
	return "ripdog"
}

// Signal return strategy signal
func (s RippleDoge) Signal() (uint8, error) {
	doge, err := gateio.Ticker("doge_usdt")
	if err != nil {
		return SIG_NONE, errors.Wrap(err, util.FuncName())
	}

	xrp, err := gateio.Ticker("xrp_usdt")
	if err != nil {
		return SIG_NONE, errors.Wrap(err, util.FuncName())
	}

	m, err := gateio.Tickers()
	if err != nil {
		return SIG_NONE, errors.Wrap(err, util.FuncName())
	}

	var rise, fall uint16
	for k, v := range m {
		if !strings.HasSuffix(k, "_usdt") {
			continue
		}
		if v.PercentChange > 0 {
			rise++
		} else {
			fall++
		}
	}
	log.Println(strconv.FormatUint(uint64(rise), 10) + "↑, " + strconv.FormatUint(uint64(fall), 10) +
		"↓, doge: " + strconv.FormatFloat(doge.PercentChange, 'f', 4, 64) +
		"%, xrp: " + strconv.FormatFloat(xrp.PercentChange, 'f', 4, 64) + "%")

	if rise > 44 && doge.PercentChange > 5 && xrp.PercentChange > 5 {
		return SIG_RISE, nil
	}
	if rise < 44 || (doge.PercentChange < -5 && xrp.PercentChange < -5) {
		return SIG_FALL, nil
	}

	return SIG_NONE, nil
}
