package gateio

import (
	"crypto/hmac"
	"crypto/sha512"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"github.com/modood/cts/util"
	"github.com/pkg/errors"
)

type (
	// Balance ...
	Balance struct {
		Result    bool
		Available map[string]string // 可用
		Locked    map[string]string // 已锁定
	}

	// Asset ...
	Asset struct {
		Balance  float64
		Pending  float64
		Total    float64
		TotalCNY float64
	}

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

	// Order ...
	Order struct {
		TradeID   string `mapstructure:"tradeID" json:"tradeID"`
		OrderID   string `mapstructure:"orderNumber" json:"orderNumber"`
		Currency  string `mapstructure:"pair" json:"pair"`
		Type      string
		Rate      string
		Amount    string
		Total     float64
		Date      string
		Timestamp string `mapstructure:"time_unix" json:"time_unix"`
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

	key    string // your api key
	secret string // your secret key
)

// CancelType ...
const (
	CancelTypeSell = 0
	CancelTypeBuy  = 1
	CancelTypeAll  = -1
)

// Init set apikey and secretkey
func Init(apikey, secretkey string) {
	key = apikey
	secret = secretkey
}

// GetPairs return all the trading pairs supported by gate.io
func GetPairs() ([]string, error) {
	bs, err := req("GET", "http://data.gate.io/api2/1/pairs", "")
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	var l []string
	err = json.Unmarshal(bs, &l)
	if err != nil {
		if err := handle(bs, nil); err != nil {
			return nil, errors.Wrap(err, util.FuncName())
		}
		return nil, errors.Wrap(err, util.FuncName())
	}

	return l, nil
}

// Tickers return pairs
func Tickers() (map[string]*Pair, error) {
	bs, err := req("GET",
		fmt.Sprintf("http://data.gate.io/api2/1/tickers"), "")
	if err := handle(bs, err); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := make(map[string]*Pair)
	for k, v := range m {
		p := Pair{}
		if err = util.Decode(v, &p); err != nil {
			return nil, errors.Wrap(err, util.FuncName())
		}

		r[k] = &p
	}

	return r, nil
}

// Ticker returns the current ticker for the selected currency,
// cached in 10 seconds.
func Ticker(currency string) (*Pair, error) {
	bs, err := req("GET",
		fmt.Sprintf("http://data.gate.io/api2/1/ticker/%s", currency), "")
	if err := handle(bs, err); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	p := Pair{}
	if err = util.Decode(m, &p); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return &p, nil
}

// MyBalance return account balances
func MyBalance() (*Balance, error) {
	bs, err := req("POST", "https://api.gate.io/api2/1/private/balances", "")
	if err := handle(bs, err); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	b := Balance{}
	if err = util.Decode(m, &b); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return &b, nil
}

// MyAsset return account assets
func MyAsset() (*Asset, error) {
	b, err := MyBalance()
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	a := Asset{}
	for k, v := range b.Available {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, errors.Wrap(err, util.FuncName())
		}

		if k == "USDT" {
			a.Balance = n
			a.Total += n
		} else {
			t, err := Ticker(strings.ToLower(k) + "_usdt")
			if err != nil {
				return nil, errors.Wrap(err, util.FuncName())
			}

			a.Total += t.Last * n
		}
	}
	for k, v := range b.Locked {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, errors.Wrap(err, util.FuncName())
		}

		if k == "USDT" {
			a.Pending += n
			a.Total += n
		} else {
			t, err := Ticker(strings.ToLower(k) + "_usdt")
			if err != nil {
				return nil, errors.Wrap(err, util.FuncName())
			}

			a.Pending += t.Last * n
			a.Total += t.Last * n
		}
	}

	r, err := Rate()
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	a.TotalCNY = r * a.Total

	return &a, nil
}

// Rate return exchange rate of USD/CNY
func Rate() (float64, error) {
	p, err := Ticker("usdt_cny")
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	return p.Last, nil
}

// Buy place order buy
func Buy(currency string, price float64, amount float64) (uint64, error) {
	params := fmt.Sprintf("currencyPair=%s&rate=%f&amount=%f",
		currency, price, amount)

	bs, err := req("POST", "https://api.gate.io/api2/1/private/buy", params)
	if err := handle(bs, err); err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	t := struct {
		OrderID uint64 `mapstructure:"orderNumber"` // 订单 ID
	}{}
	if err := util.Decode(m, &t); err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	return t.OrderID, nil
}

// Sell place order sell
func Sell(currency string, price float64, amount float64) (uint64, error) {
	params := fmt.Sprintf("currencyPair=%s&rate=%f&amount=%f",
		currency, price, amount)

	bs, err := req("POST", "https://api.gate.io/api2/1/private/sell", params)
	if err := handle(bs, err); err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	t := struct {
		OrderID uint64 `mapstructure:"orderNumber"` // 订单 ID
	}{}
	if err := util.Decode(m, &t); err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	return t.OrderID, nil
}

// Cancel cancel all orders
func Cancel(currency string, cancelType int8) error {
	params := fmt.Sprintf("currencyPair=%s&type=%d", currency, cancelType)
	url := "https://api.gate.io/api2/1/private/cancelAllOrders"

	bs, err := req("POST", url, params)
	err = handle(bs, err)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

// LatestOrder return latest order of my last 24h trades
func LatestOrder(currency string) (*Order, error) {
	url := "https://api.gate.io/api2/1/private/tradeHistory"
	bs, err := req("POST", url, "currencyPair="+currency)
	if err := handle(bs, err); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Trades []*Order }{}

	err = json.Unmarshal(bs, &r)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	if len(r.Trades) != 0 {
		return r.Trades[0], nil
	}

	return &Order{}, nil
}

// OpenOrderLen return count of open orders
func OpenOrderLen() (int, error) {
	bs, err := req("POST", "https://api.gate.io/api2/1/private/openOrders", "")
	if err := handle(bs, err); err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	r := struct {
		Orders []interface{} `json:"orders"`
	}{}

	err = json.Unmarshal(bs, &r)
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	return len(r.Orders), nil
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

func sign(content string) (string, error) {
	h := hmac.New(sha512.New, []byte(secret))
	_, err := h.Write([]byte(content))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func req(method string, url string, param string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, strings.NewReader(param))
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	s, err := sign(param)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("key", key)
	req.Header.Set("sign", s)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return bs, nil
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

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
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
