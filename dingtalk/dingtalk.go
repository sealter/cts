package dingtalk

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

var (
	token string
)

func Init(accessToken string) {
	token = accessToken
}

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
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)

	return err
}
