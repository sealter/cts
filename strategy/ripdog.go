package strategy

import "github.com/modood/cts/gateio"

// RippleDoge is a simple strategy refers to ripple and doge
type RippleDoge struct{}

func (s RippleDoge) Name() string {
	return "ripdog"
}

func (s RippleDoge) Signal() (uint8, error) {
	doge, err := gateio.Ticker("doge_usdt")
	if err != nil {
		return SIG_NONE, err
	}

	xrp, err := gateio.Ticker("xrp_usdt")
	if err != nil {
		return SIG_NONE, err
	}

	if doge.PercentChange > 5 && xrp.PercentChange > 5 {
		return SIG_RISE, nil
	}
	if doge.PercentChange < -5 && xrp.PercentChange < -5 {
		return SIG_FALL, nil
	}

	return SIG_NONE, nil
}
