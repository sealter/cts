package dingtalk

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/modood/cts/util"
	"github.com/pkg/errors"
)

var (
	token string
)

// Init init access token of dingtalk group chat robot
func Init(accessToken string) {
	token = accessToken
}

// Push send notification
func Push(text string, isAtAll bool) error {
	client := &http.Client{Timeout: time.Duration(time.Second * 3)}

	url := "https://oapi.dingtalk.com/robot/send?access_token=" + token

	if isAtAll {
		text += "\n"
	}

	params := []byte(`{
		"msgtype": "text",
		"text": {
			"content": "` + text + `"
		},
		"at": {
			"atMobiles": [],
			"isAtAll": ` + strconv.FormatBool(isAtAll) + `
		}
	}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(params))
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	req.Header.Set("Content-Type", "application/json")

	var retry int
t:
	resp, err := client.Do(req)
	if err != nil {
		if err, ok := err.(net.Error); (ok && err.Timeout()) ||
			strings.Contains(err.Error(), "connection reset by peer") {
			if retry++; retry < 3 {
				goto t
			}
		}
		return errors.Wrap(err, util.FuncName())
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	r := struct {
		Code    int32  `json:"errcode"`
		Message string `json:"errmsg"`
	}{}

	err = json.Unmarshal(bs, &r)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	if r.Code != 0 || r.Message != "ok" {
		err = errors.New(string(bs))
		return errors.Wrap(err, util.FuncName())
	}

	return nil
}
