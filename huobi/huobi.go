package huobi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"github.com/modood/cts/dingtalk"
	"github.com/modood/cts/util"
	"github.com/pkg/errors"
)

type (
	// Symbol ...
	Symbol struct {
		Name            string
		BaseCurrency    string `mapstructure:"base-currency" json:"base-currency"`
		QuoteCurrency   string `mapstructure:"quote-currency" json:"quote-currency"`
		PricePrecision  int    `mapstructure:"price-precision" json:"price-precision"`
		AmountPrecision int    `mapstructure:"amount-precision" json:"amount-precision"`
		SymbolPartition string `mapstructure:"symbol-partition" json:"symbol-partition"`
	}

	// Account ...
	Account struct {
		ID       uint64
		Type     string
		State    string
		Symbol   string
		FlPrice  float64 `mapstructure:"fl-price" json:"fl-price"`
		FlType   string  `mapstructure:"fl-type" json:"fl-type"`
		RiskRate float64 `mapstructure:"risk-rate" json:"risk-rate"`
		List     []struct {
			Currency string
			Type     string
			Balance  float64
		}
	}

	// Carry ...
	Carry struct {
		Trade                float64
		Frozen               float64
		TransferOutAvailable float64
		LoanAvailable        float64
		Loan                 float64
		Interest             float64
	}

	// Limit ...
	Limit struct {
		BuyGT  float64 `mapstructure:"market-buy-order-must-greater-than" json:"market-buy-order-must-greater-than"`
		BuyLT  float64 `mapstructure:"market-buy-order-must-less-than" json:"market-buy-order-must-less-than"`
		SellGT float64 `mapstructure:"market-sell-order-must-greater-than" json:"market-sell-order-must-greater-than"`
		SellLT float64 `mapstructure:"market-sell-order-must-less-than" json:"market-sell-order-must-less-than"`
	}

	// OpenOrder ...
	OpenOrder struct {
		ID              uint64
		AccountID       uint64 `mapstructure:"account-id" json:"account-id"`
		Source          string
		Type            string
		State           string
		Symbol          string
		Amount          float64
		Price           float64
		FieldAmount     float64 `mapstructure:"field-amount" json:"field-amount"`
		FieldCashAmount float64 `mapstructure:"field-cash-amount" json:"field-cash-amount"`
		FieldFees       float64 `mapstructure:"field-fees" json:"field-fees"`
		CreatedAt       uint64  `mapstructure:"created-at" json:"created-at"`
		FinishedAt      uint64  `mapstructure:"finished-at" json:"finished-at"`
		CanceledAt      uint64  `mapstructure:"canceled-at" json:"canceled-at"`
	}

	// Order ...
	Order struct {
		ID        uint64
		OrderID   uint64 `mapstructure:"order-id" json:"order-id"`
		MatchID   uint64 `mapstructure:"match-id" json:"match-id"`
		Source    string
		Type      string
		Symbol    string
		Price     float64
		Amount    float64 `mapstructure:"filled-amount" json:"filled-amount"`
		Fees      float64 `mapstructure:"filled-fees" json:"filled-fees"`
		Points    float64 `mapstructure:"filled-points" json:"filled-points"`
		CreatedAt uint64  `mapstructure:"created-at" json:"created-at"`
	}

	// BorrowOrder ...
	BorrowOrder struct {
		ID              uint64
		State           string
		UserID          uint64 `mapstructure:"user-id" json:"user-id"`
		AccountID       uint64 `mapstructure:"account-id" json:"account-id"`
		Symbol          string
		Currency        string
		LoanAmount      float64 `mapstructure:"loan-amount" json:"loan-amount"`
		LoanBalance     float64 `mapstructure:"loan-balance" json:"loan-balance"`
		InterestAmount  float64 `mapstructure:"interest-amount" json:"interest-amount"`
		InterestBalance float64 `mapstructure:"interest-balance" json:"interest-balance"`
		InterestRate    float64 `mapstructure:"interest-rate" json:"interest-rate"`
		CreatedAt       uint64  `mapstructure:"created-at" json:"created-at"`
		UpdatedAt       uint64  `mapstructure:"updated-at" json:"updated-at"`
		AccruedAt       uint64  `mapstructure:"accrued-at" json:"accrued-at"`
	}

	huobiError struct {
		Status string
		// Error code:
		// base-symbol-error                            交易对不存在
		// base-currency-error                          币种不存在
		// base-date-error                              错误的日期格式
		// account-transfer-balance-insufficient-error  余额不足无法冻结
		// bad-argument                                 无效参数
		// api-signature-not-valid                      API签名错误
		// gateway-internal-error                       系统繁忙，请稍后再试
		// security-require-assets-password             需要输入资金密码
		// audit-failed                                 下单失败
		// ad-ethereum-addresss                         请输入有效的以太坊地址
		// order-accountbalance-error                   账户余额不足
		// order-limitorder-price-error                 限价单下单价格超出限制
		// order-limitorder-amount-error                限价单下单数量超出限制
		// order-orderprice-precision-error             下单价格超出精度限制
		// order-orderamount-precision-error            下单数量超过精度限制
		// order-marketorder-amount-error               下单数量超出限制
		// order-queryorder-invalid                     查询不到此条订单
		// order-orderstate-error                       订单状态错误
		// order-datelimit-error                        查询超出时间限制
		// order-update-error                           订单更新出错
		// bad-request                                  错误请求
		// invalid-parameter                            参数错
		// invalid-command                              指令错
		Code    string `mapstructure:"err-code" json:"err-code"`
		Message string `mapstructure:"err-msg" json:"err-msg"`
	}
)

var (
	key    string // your api key
	secret string // your secret key

	errInvalidSymbol     = errors.New("invalid symbol name, A valid name should look like: btc_usdt")
	errInvalidCurrency   = errors.New("invalid currency")
	errUnsupportedSymbol = errors.New("unsupported symbol")
	errNoMarginAccount   = errors.New("no margin account")
	errUnkownTradeType   = errors.New("unknown trade type, it should be `BUY` or `SELL`")

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// Init set apikey and secretkey
func Init(apikey, secretkey string) {
	key = apikey
	secret = secretkey
}

// Symbols return all support symbol
func Symbols() ([]Symbol, error) {
	m, err := req("GET", "https://api.huobi.pro/v1/common/symbols", nil)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data []Symbol }{}
	if err = util.Decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return r.Data, nil
}

// OrderDetail return order detail by ID
func OrderDetail(ID uint64) (*OpenOrder, error) {
	m, err := req("GET", "https://api.huobi.pro/v1/order/orders/"+
		strconv.FormatUint(ID, 10), nil)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data OpenOrder }{}
	if err = util.Decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return &r.Data, nil
}

// NewSymbol return new symbol
func NewSymbol(name string) (*Symbol, error) {
	n := strings.Split(name, "_")
	if len(n) != 2 {
		return nil, errors.Wrap(errInvalidSymbol, util.FuncName())
	}

	ss, err := Symbols()
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	var s *Symbol
	for _, v := range ss {
		if n[0] == v.BaseCurrency && n[1] == v.QuoteCurrency {
			s = &v
			s.Name = v.BaseCurrency + v.QuoteCurrency
			break
		}
	}
	if s == nil {
		return nil, errors.Wrap(errUnsupportedSymbol, util.FuncName())
	}

	return s, nil
}

// Limit return trade limit of a symbol
func (s *Symbol) Limit() (*Limit, error) {
	m, err := req("GET", "https://api.huobi.pro/v1/common/exchange",
		map[string]string{"symbol": s.Name})
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data Limit }{}
	if err = util.Decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return &r.Data, nil
}

// Account return margin account
func (s *Symbol) Account() (*Account, error) {
	m, err := req("GET", "https://api.huobi.pro/v1/margin/accounts/balance",
		map[string]string{"symbol": s.Name})
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data []Account }{}
	if err = util.Decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	if len(r.Data) == 0 {
		return nil, errors.Wrap(errNoMarginAccount, util.FuncName())
	}

	return &r.Data[0], nil
}

// Carry return balance of specific currency
func (s *Symbol) Carry(currency string) (*Carry, error) {
	a, err := s.Account()
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	c := Carry{}
	for _, v := range a.List {
		if v.Currency != currency {
			continue
		}
		switch v.Type {
		case "trade":
			c.Trade = v.Balance
		case "frozen":
			c.Frozen = v.Balance
		case "transfer-out-available":
			c.TransferOutAvailable = v.Balance
		case "loan-available":
			c.LoanAvailable = v.Balance
		case "loan":
			c.Loan = v.Balance
		case "interest":
			c.Interest = v.Balance
		}
	}
	return &c, nil
}

// Orders return finished orders
func (s *Symbol) Orders() ([]Order, error) {
	m, err := req("GET", "https://api.huobi.pro/v1/order/matchresults",
		map[string]string{"symbol": s.Name})
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data []Order }{}
	if err := util.Decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return r.Data, nil
}

// OpenOrders return pendding orders
func (s *Symbol) OpenOrders(state string) ([]OpenOrder, error) {
	if state == "" {
		state = "pre-submitted,submitted,partial-filled"
	}

	m, err := req("GET", "https://api.huobi.pro/v1/order/orders",
		map[string]string{
			"symbol": s.Name,
			"states": state,
		})
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data []OpenOrder }{}
	if err := util.Decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return r.Data, nil
}

// BorrowOrders return borrow orders
func (s *Symbol) BorrowOrders(state string) ([]BorrowOrder, error) {
	m, err := req("GET", "https://api.huobi.pro/v1/margin/loan-orders",
		map[string]string{
			"symbol": s.Name,
			"states": state,
		})
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data []BorrowOrder }{}
	if err := util.Decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return r.Data, nil
}

// BorrowAvailable return available amount to borrow
func (s *Symbol) BorrowAvailable(currency string) (float64, error) {
	if currency != s.BaseCurrency && currency != s.QuoteCurrency {
		return 0, errors.Wrap(errInvalidCurrency, util.FuncName())
	}

	a, err := s.Account()
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}
	for _, v := range a.List {
		if v.Type == "loan-available" && v.Currency == currency {
			return v.Balance, nil
		}
	}
	return 0, nil
}

// Borrow borrow money
func (s *Symbol) Borrow(currency string, amount float64) error {
	if currency != s.BaseCurrency && currency != s.QuoteCurrency {
		return errors.Wrap(errInvalidCurrency, util.FuncName())
	}

	_, err := req("POST", "https://api.huobi.pro/v1/margin/orders",
		map[string]string{
			"symbol":   s.Name,
			"currency": currency,
			"amount":   floor(amount, 3),
		})
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	msg := fmt.Sprintf("%s\n类型：%s\n品种：%s\n数量：%.4f %s",
		time.Now().Format("2006-01-02 15:04:05"),
		"borrow", s.Name, amount, currency)
	err = dingtalk.Push(msg)
	if err != nil {
		log.Println(err)
	}

	return nil
}

// Repay repay all debt
func (s *Symbol) Repay(currency string) error {
	if currency != s.BaseCurrency && currency != s.QuoteCurrency {
		return errors.Wrap(errInvalidCurrency, util.FuncName())
	}

	bos, err := s.BorrowOrders("accrual")
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	var errs []string
	for _, v := range bos {
		if v.Currency != currency {
			continue
		}

		_, err := req("POST", "https://api.huobi.pro/v1/margin/orders/"+
			strconv.FormatUint(v.ID, 10)+"/repay",
			map[string]string{
				"amount": floor(v.LoanAmount+v.InterestAmount, 8),
			})
		if err != nil {
			errs = append(errs, err.Error()+"(ID: "+strconv.FormatUint(v.ID, 10)+")")
			continue
		}

		msg := fmt.Sprintf("%s\n类型：%s\n品种：%s\n数量：%.4f %s\n利息：%.6f %s",
			time.Now().Format("2006-01-02 15:04:05"),
			"repay", s.Name, v.LoanAmount, currency, v.InterestAmount, currency)
		err = dingtalk.Push(msg)
		if err != nil {
			log.Println(err)
		}
	}
	if len(errs) != 0 {
		return errors.Wrap(errors.New(strings.Join(errs, ";")), util.FuncName())
	}
	return nil
}

// Trade place new margin order
func (s *Symbol) Trade(cmd string, amount float64) error {
	a, err := s.Account()
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	params := map[string]string{
		"account-id": strconv.FormatUint(a.ID, 10),
		"source":     "margin-api",
		"symbol":     s.Name,
		"amount":     floor(amount, s.AmountPrecision),
	}

	switch cmd {
	case "BUY":
		params["type"] = "buy-market"
	case "SELL":
		params["type"] = "sell-market"
	/* testing */
	case "TESTBUY":
		params["price"] = "1"
		params["type"] = "buy-limit"
	case "TESTSELL":
		params["price"] = "100000"
		params["type"] = "sell-limit"
	default:
		return errors.Wrap(errUnkownTradeType, util.FuncName())
	}

	m, err := req("POST", "https://api.huobi.pro/v1/order/orders/place", params)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	r := struct{ Data uint64 }{}
	if err := util.Decode(m, &r); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	time.Sleep(time.Second * 5) // await until order state changed: submitted => filled
	o, err := OrderDetail(r.Data)
	if err := util.Decode(m, &r); err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	msg := fmt.Sprintf("%s\n订单：%d\n状态：%s\n类型：%s\n品种：%s\n价格：$%.2f\n数量：$%.2f",
		time.Now().Format("2006-01-02 15:04:05"), o.ID, o.State,
		strings.ToLower(cmd), o.Symbol, o.FieldCashAmount/o.FieldAmount, o.FieldCashAmount)

	err = dingtalk.Push(msg)
	if err != nil {
		log.Println(err)
	}

	return nil
}

// CancelAll cancel all open orders
func (s *Symbol) CancelAll() error {
	oos, err := s.OpenOrders("")
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	var errs []string
	for _, v := range oos {
		_, err := req("POST", "https://api.huobi.pro/v1/order/orders/"+
			strconv.FormatUint(v.ID, 10)+"/submitcancel", nil)
		if err != nil {
			errs = append(errs, err.Error()+"(ID: "+strconv.FormatUint(v.ID, 10)+")")
			continue
		}
	}
	if len(errs) != 0 {
		return errors.Wrap(errors.New(strings.Join(errs, ";")), util.FuncName())
	}
	return nil
}

// AllIn all in
func (s *Symbol) AllIn(cmd string, isMargin bool) error {
	bc, qc := s.BaseCurrency, s.QuoteCurrency
	switch cmd {
	case "BUY": // do nothing
	case "SELL":
		bc, qc = qc, bc
	default:
		return errors.Wrap(errUnkownTradeType, util.FuncName())
	}

	err := s.CancelAll()
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

t:
	c, err := s.Carry(qc)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}
	if isMargin && c.LoanAvailable > 0 {
		bos, err := s.BorrowOrders("accrual")
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}
		if len(bos) == 0 {
			err = s.Borrow(qc, c.LoanAvailable)
			if err != nil {
				return errors.Wrap(err, util.FuncName())
			}
			goto t
		}
	}

	// check trade amount limit
	l, err := s.Limit()
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}
	if cmd == "BUY" {
		if c.Trade < l.BuyGT {
			return nil
		} else if c.Trade > l.BuyLT {
			c.Trade = l.BuyLT
		}
	}
	if cmd == "SELL" {
		if c.Trade < l.SellGT {
			return nil
		} else if c.Trade > l.SellLT {
			c.Trade = l.SellLT
		}
	}

	err = s.Trade(cmd, c.Trade)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	err = s.Repay(bc)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}

func sign(content string) (string, error) {
	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(content))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func req(method, address string, params map[string]string) (map[string]interface{}, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	host := u.Hostname()
	path := u.EscapedPath()

	compute := map[string]string{
		"AccessKeyId":      key,
		"SignatureMethod":  "HmacSHA256",
		"SignatureVersion": "2",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05"),
	}

	var ctype, signature string
	var reader io.Reader
	switch strings.ToUpper(method) {
	case "GET":
		for k, v := range params {
			compute[k] = v
		}
		ctype = "application/x-www-form-urlencoded"
	default:
		ctype = "application/json"

		bs, err := json.Marshal(params)
		if err != nil {
			return nil, errors.Wrap(err, util.FuncName())
		}
		reader = bytes.NewBuffer(bs)
	}

	query := querystring(compute)
	signature, err = sign(method + "\n" + host + "\n" + path + "\n" + query)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	// huobi get parameters must be passing by querystring
	address += "?" + query + "&Signature=" + url.QueryEscape(signature)

	client := &http.Client{}

	req, err := http.NewRequest(method, address, reader)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	req.Header.Set("Content-Type", ctype)

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
	if err := handle(bs, err); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return m, nil
}

func querystring(m map[string]string) string {
	l := len(m)

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	q := make([]string, l)
	for i, k := range keys {
		q[i] = url.QueryEscape(k) + "=" + url.QueryEscape(m[k])
	}

	return strings.Join(q, "&")
}

func floor(f float64, prec int) string {
	i := math.Pow10(prec)
	return strconv.FormatFloat(math.Floor(f*i)/i, 'f', -1, 64)
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

	e := huobiError{}

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

	if e.Status != "ok" {
		err = fmt.Errorf("Code: %s, %s", e.Code, e.Message)
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}
