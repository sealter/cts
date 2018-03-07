package gateio

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"github.com/modood/cts/util"
	"github.com/pkg/errors"
)

type (
	// Pair ...
	Pair struct {
		Result        bool
		PercentChange float64 // 涨跌百分比
		Last          float64 // 最新成交价
		LowestAsk     float64 // 卖方最低价
		HighestBid    float64 // 买方最高价
		BaseVolume    float64 // 交易量
		QuoteVolume   float64 // 兑换货币交易量
		High24hr      float64 // 24 小时最高价
		Low24hr       float64 // 24 小时最低价
	}

	gateioError struct {
		// Response:
		// true		success
		// false	fail
		Result bool

		// Error code:
		// 0	Success
		// 1	Invalid request
		// 2	Invalid version
		// 3	Invalid request
		// 4	Too many attempts
		// 5,6	Invalid sign
		// 7	Currency is not supported
		// 8,9	Currency is not supported
		// 10	Verified failed
		// 11	Obtaining address failed
		// 12	Empty params
		// 13	Internal error, please report to administrator
		// 14	Invalid user
		// 15	Cancel order too fast, please wait 1 min and try again
		// 16	Invalid order id or order is already closed
		// 17	Invalid orderid
		// 18	Invalid amount
		// 19	Not permitted or trade is disabled
		// 20	Your order size is too small
		// 21	You don't have enough fund
		Code    int32
		Message string
	}
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// Tickers return pairs
func Tickers() (map[string]*Pair, error) {
	m, err := get("https://data.gate.io/api2/1/tickers")
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := make(map[string]*Pair)
	for k, v := range m {
		p := Pair{}
		if err = util.Decode(v.(map[string]interface{}), &p); err != nil {
			return nil, errors.Wrap(err, util.FuncName())
		}

		r[k] = &p
	}

	return r, nil
}

// Ticker returns ticker for the selected symbol
func Ticker(symbol string) (*Pair, error) {
	m, err := get("https://data.gate.io/api2/1/ticker/" + symbol)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	p := Pair{}
	if err = util.Decode(m, &p); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return &p, nil
}

// Rate return exchange rate of USD/CNY
func Rate() (float64, error) {
	p, err := Ticker("usdt_cny")
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	return p.Last, nil
}

// Trend return market trend
func Trend() (rise, fall uint16, err error) {
	m, err := Tickers()
	if err != nil {
		return 0, 0, errors.Wrap(err, util.FuncName())
	}

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
	return rise, fall, nil
}

func get(address string) (map[string]interface{}, error) {
	var retry int
t:
	resp, err := http.Get(address)
	if err != nil {
		if retry++; retry < 3 {
			time.Sleep(time.Second * 10) // retry 10 seconds later
			goto t
		}
		return nil, errors.Wrap(err, util.FuncName())
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	bs, err := ioutil.ReadAll(resp.Body)
	if err := handle(bs, err); err != nil {
		if retry++; retry < 3 {
			time.Sleep(time.Second * 10) // retry 10 seconds later
			goto t
		}
		return nil, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		if retry++; retry < 3 {
			time.Sleep(time.Second * 10) // retry 10 seconds later
			goto t
		}
		return nil, errors.Wrap(err, util.FuncName())
	}

	return m, nil
}

func handle(bs []byte, err error) error {
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	e := gateioError{}

	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           &e,
		})
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	err = decoder.Decode(m)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	if e.Code != 0 {
		err = fmt.Errorf("Code: %d, %s", e.Code, e.Message)
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}
