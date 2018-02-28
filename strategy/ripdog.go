package strategy

import (
	"log"
	"strconv"

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
		return SigNone, errors.Wrap(err, util.FuncName())
	}

	xrp, err := gateio.Ticker("xrp_usdt")
	if err != nil {
		return SigNone, errors.Wrap(err, util.FuncName())
	}

	rise, fall, err := gateio.Trend()
	if err != nil {
		return SigNone, errors.Wrap(err, util.FuncName())
	}

	log.Println(strconv.FormatUint(uint64(rise), 10) + "↑, " + strconv.FormatUint(uint64(fall), 10) +
		"↓, doge: " + strconv.FormatFloat(doge.PercentChange, 'f', 4, 64) +
		"%, xrp: " + strconv.FormatFloat(xrp.PercentChange, 'f', 4, 64) + "%")

	if rise > 44 && doge.PercentChange > 5 && xrp.PercentChange > 5 {
		return SigRise, nil
	}
	if rise < 44 || (doge.PercentChange < -5 && xrp.PercentChange < -5) {
		return SigFall, nil
	}

	return SigNone, nil
}
