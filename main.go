package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Login    []Login `json:"login"`
	PushPlus string  `json:"pushPlus"`
}

type Login struct {
	URL      string `json:"url"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Cookie   string `json:"cookie"`
}

type Ret struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}

func main() {
	sp()
}

func sp() {
	config := &Config{}
	osConfig := os.Getenv("CONFIG")
	json.Unmarshal([]byte(osConfig), config)

	content := ""
	for _, gin := range config.Login {
		// 登陆
		cookie := login(gin)
		if cookie == "" {
			cookie = gin.Cookie
		}
		// 签到
		result := sign(gin.URL, cookie)
		content += gin.URL + "\n" + result + "\n\n"
	}

	// 通知
	sendPushPlus(config.PushPlus, "SP 签到", content)
}

func login(gin Login) string {
	path := gin.URL + "/auth/login"
	params := url.Values{}
	params.Add("email", gin.Email)
	params.Add("passwd", gin.Password)
	resp, _ := http.PostForm(path, params)
	defer resp.Body.Close()

	cookies := ""
	for _, cookie := range resp.Cookies() {
		cook := strings.Split(cookie.String(), ";")[0]
		cookies += cook + "; "
	}
	cookies = strings.TrimRight(cookies, " ")
	return cookies
}

func sign(url, cookie string) string {
	url = url + "/user/checkin"
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Cookie", cookie)
	client := http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	ret := &Ret{}
	err := json.Unmarshal(body, ret)
	if err != nil {
		return "Cookie 已失效！"
	}

	return ret.Msg
}

func sendPushPlus(token, title, content string) {
	url := "https://www.pushplus.plus/send"
	ma := make(map[string]interface{})
	ma["token"] = token
	ma["title"] = title
	ma["content"] = content
	js, _ := json.Marshal(ma)
	param := bytes.NewReader(js)

	req, _ := http.NewRequest("POST", url, param)
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
}
