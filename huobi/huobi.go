package huobi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
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

	"github.com/mitchellh/mapstructure"
	"github.com/modood/cts/dingtalk"
	"github.com/modood/cts/util"
	"github.com/pkg/errors"
)

type (
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
	//json = jsoniter.ConfigCompatibleWithStandardLibrary

	key    string // your api key
	secret string // your secret key
)

// Init set apikey and secretkey
func Init(apikey, secretkey string) {
	key = apikey
	secret = secretkey
}

// MarginAccount return margin account information
func MarginAccount(currency string) (*Account, error) {
	params := make(map[string]string)

	currency = hCurrency(currency)
	if currency != "" {
		params["symbol"] = currency
	}

	bs, err := req("GET", "https://api.huobi.pro/v1/margin/accounts/balance", params)
	if err := handle(bs, err); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	r := struct {
		Data []Account
	}{}
	if err = util.Decode(m, &r); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	if len(r.Data) == 0 {
		return nil, errors.Wrap(errors.New("no margin account"), util.FuncName())
	}

	return &r.Data[0], nil
}

// Orders return finished orders
func Orders(currency string) ([]Order, error) {
	bs, err := req("GET", "https://api.huobi.pro/v1/order/matchresults",
		map[string]string{
			"symbol": hCurrency(currency),
		})
	if err := handle(bs, err); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	t := struct {
		Orders []Order `mapstructure:"data"`
	}{}
	if err := util.Decode(m, &t); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return t.Orders, nil
}

// OpenOrders return pendding orders
func OpenOrders(currency string) ([]OpenOrder, error) {
	bs, err := req("GET", "https://api.huobi.pro/v1/order/orders",
		map[string]string{
			"symbol": hCurrency(currency),
			"states": "pre-submitted,submitted,partial-filled",
		})
	if err := handle(bs, err); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	t := struct {
		Orders []OpenOrder `mapstructure:"data"`
	}{}
	if err := util.Decode(m, &t); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return t.Orders, nil
}

// BorrowOrders return borrow orders
func BorrowOrders(currency string) ([]BorrowOrder, error) {
	bs, err := req("GET", "https://api.huobi.pro/v1/margin/loan-orders",
		map[string]string{
			"symbol": hCurrency(currency),
		})
	if err := handle(bs, err); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	t := struct {
		Orders []BorrowOrder `mapstructure:"data"`
	}{}
	if err := util.Decode(m, &t); err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	return t.Orders, nil
}

// MarginTrade place new margin order
func MarginTrade(cmd, currency string, amount float64) (uint64, error) {
	a, err := MarginAccount(currency)
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	params := map[string]string{
		"account-id": strconv.FormatUint(a.ID, 10),
		"amount":     strconv.FormatFloat(amount, 'f', -1, 64),
		"source":     "margin-api",
		"symbol":     hCurrency(currency),
	}

	switch cmd {
	case "BUY":
		params["type"] = "buy-market"
	case "SELL":
		params["type"] = "sell-market"
	case "TEST":
		params["price"] = "100"
		params["type"] = "buy-limit"
	default:
		return 0, errors.Wrap(errors.New("unknown trade type"), util.FuncName())
	}

	bs, err := req("POST", "https://api.huobi.pro/v1/order/orders/place", params)
	if err := handle(bs, err); err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	t := struct {
		ID string `mapstructure:"data"`
	}{}
	if err := util.Decode(m, &t); err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	return strconv.ParseUint(t.ID, 10, 64)
}

// BorrowAvailable return available amount to borrow
func BorrowAvailable(currency, symbol string) (float64, error) {
	a, err := MarginAccount(currency)
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}
	for _, v := range a.List {
		if v.Type == "loan-available" && v.Currency == symbol {
			return v.Balance, nil
		}
	}
	return 0, nil
}

// Borrow borrow money
func Borrow(currency, symbol string, amount float64) (uint64, error) {
	params := map[string]string{
		"symbol":   hCurrency(currency),
		"currency": symbol, // lol, funny ^_^. currency => symbol, symbol => currency
		"amount":   floor(amount, 3),
	}

	bs, err := req("POST", "https://api.huobi.pro/v1/margin/orders", params)
	if err := handle(bs, err); err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	t := struct {
		ID string `mapstructure:"data"`
	}{}
	if err := util.Decode(m, &t); err != nil {
		return 0, errors.Wrap(err, util.FuncName())
	}

	return strconv.ParseUint(t.ID, 10, 64)
}

// Repay all debt
func Repay(currency, symbol string) error {
	bos, err := BorrowOrders(currency)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	var errs []string
	for _, v := range bos {
		if v.Currency != symbol || v.State != "accrual" {
			continue
		}
		bs, err := req("POST", "https://api.huobi.pro/v1/margin/orders/"+strconv.FormatUint(v.ID, 10)+"/repay",
			map[string]string{
				"amount": floor(v.LoanAmount+v.InterestAmount, 8),
			})
		if err := handle(bs, err); err != nil {
			errs = append(errs, err.Error()+"(ID: "+strconv.FormatUint(v.ID, 10)+")")
			continue
		}

		msg := fmt.Sprintf("%s\n类型：%s\n品种：%s\n数量：%.4f %s",
			time.Now().Format("2006-01-02 15:04:05"), "repay", currency, v.LoanAmount+v.InterestAmount, symbol)
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

// CancelAll cancel all open orders
func CancelAll(currency string) error {
	oos, err := OpenOrders(currency)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	var errs []string
	for _, v := range oos {
		bs, err := req("POST", "https://api.huobi.pro/v1/order/orders/"+
			strconv.FormatUint(v.ID, 10)+"/submitcancel", map[string]string{})
		if err := handle(bs, err); err != nil {
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
func AllIn(cmd, currency string, isMargin bool) error {
	buySymbol, sellSymbol, err := symbols(currency)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}
	switch cmd {
	case "BUY": // do nothing
	case "SELL":
		buySymbol, sellSymbol = sellSymbol, buySymbol
	default:
		return errors.Wrap(errors.New("invalid cmd, it should be `BUY` or `SELL`"), util.FuncName())
	}

	err = CancelAll(currency)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

t:
	c, err := carry(currency, sellSymbol)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}
	if isMargin && c.LoanAvailable > 0 {
		_, err = Borrow(currency, sellSymbol, c.LoanAvailable)
		if err != nil {
			return errors.Wrap(err, util.FuncName())
		}

		msg := fmt.Sprintf("%s\n类型：%s\n品种：%s\n数量：%.4f %s",
			time.Now().Format("2006-01-02 15:04:05"), "borrow", currency, c.LoanAvailable, sellSymbol)
		err = dingtalk.Push(msg)
		if err != nil {
			log.Println(err)
		}

		goto t
	}

	if c.Trade <= 0.00000001 {
		return nil
	}

	_, err = MarginTrade(cmd, currency, c.Trade)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	msg := fmt.Sprintf("%s\n类型：%s\n品种：%s\n数量：%.4f %s",
		time.Now().Format("2006-01-02 15:04:05"), strings.ToLower(cmd), currency, c.Trade, sellSymbol)
	err = dingtalk.Push(msg)
	if err != nil {
		log.Println(err)
	}

	err = Repay(currency, buySymbol)
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

func req(method, address string, params map[string]string) ([]byte, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}
	host := u.Hostname()
	path := u.EscapedPath()

	m := map[string]string{
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
			m[k] = v
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

	query := querystring(m)
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

	return bs, nil
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

func hCurrency(currency string) string {
	return strings.Replace(currency, "_", "", -1)
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

func carry(currency, symbol string) (*Carry, error) {
	a, err := MarginAccount(currency)
	if err != nil {
		return nil, errors.Wrap(err, util.FuncName())
	}

	c := Carry{}
	for _, v := range a.List {
		if v.Currency != symbol {
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

func symbols(currency string) (string, string, error) {
	s := strings.Split(currency, "_")
	if len(s) != 2 {
		return "", "", errors.Wrap(errors.New("invalid currency, A valid currency should look like: btc_usdt"), util.FuncName())
	}

	return s[0], s[1], nil
}
