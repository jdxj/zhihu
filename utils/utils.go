package utils

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
)

// StringToCookies 将给定的 cookie 字符串转换成 http.Cookie,
// domain 是 http.Cookie 所必须的.
func StringToCookies(cookiesStr, domain string) ([]*http.Cookie, error) {
	if domain == "" {
		return nil, fmt.Errorf("invaild domain")
	}

	// 过滤引号
	cookiesStr = strings.ReplaceAll(cookiesStr, `"`, ``)
	// 过滤空格
	cookiesStr = strings.ReplaceAll(cookiesStr, ` `, ``)
	// 划分
	cookiesParts := strings.Split(cookiesStr, ";")

	var cookies []*http.Cookie
	for _, part := range cookiesParts {
		idx := strings.Index(part, "=")
		if idx < 0 {
			logs.Warn("%s not found '=' in cookie part: %s", "zhihu", part)
			continue
		}
		k := part[:idx]
		v := part[idx+1:]

		cookie := &http.Cookie{
			Name:     k,
			Value:    v,
			Path:     "/",
			Domain:   domain,
			Expires:  time.Now().Add(time.Hour * 24 * 365), // 一年后过期
			Secure:   false,
			HttpOnly: false,
		}
		cookies = append(cookies, cookie)
	}

	if len(cookies) == 0 {
		return nil, fmt.Errorf("invalid cookie")
	}
	return cookies, nil
}

func NewRequestWithUserAgent(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36")
	return req, err
}

func Pause5s() {
	logs.Debug("pause 5s, start: %d", time.Now().Unix())
	pause(5 * time.Second)
	logs.Debug("pause 5s, end  : %d", time.Now().Unix())
}

func pause(dur time.Duration) {
	timer := time.NewTimer(dur)
	defer timer.Stop()

	select {
	case <-timer.C:
	}
}
