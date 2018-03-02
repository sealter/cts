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
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/modood/cts/util"
	"github.com/pkg/errors"
)

type (
	// Account ...
	Account struct {
		ID       uint32
		Type     string
		State    string
		Balances []Available `mapstructure:"list" json:"list"`
	}

	// Available ...
	Available struct {
		Currency string
		Type     string // 可用
		Balance  string // 已锁定
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
