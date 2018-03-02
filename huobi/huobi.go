package huobi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/modood/cts/util"
	"github.com/pkg/errors"
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
