package dingtalk

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

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
func Push(text string) error {
	client := &http.Client{}

	url := "https://oapi.dingtalk.com/robot/send?access_token=" + token

	params := []byte(`{
		"msgtype": "text",
		"text": {
			"content": "` + text + `"
		},
		"at": {
			"atMobiles": [],
			"isAtAll": false
		}
	}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(params))
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, util.FuncName())
	}

	defer resp.Body.Close()

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
