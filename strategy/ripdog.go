package strategy

import (
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

	if doge.PercentChange > 5 && xrp.PercentChange > 5 {
		return SIG_RISE, nil
	}
	if doge.PercentChange < -5 && xrp.PercentChange < -5 {
		return SIG_FALL, nil
	}

	return SIG_NONE, nil
}
