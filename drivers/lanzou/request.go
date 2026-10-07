package lanzou

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/AlliotTech/openalist/drivers/base"
	"github.com/go-resty/resty/v2"
)

const loginURL = "https://up.woozooo.com/mlogin.php"
const diskURL = "https://pc.woozooo.com/mydisk.php"
const challengeCookie = "acw_sc__v2"

// Account and challenge cookies must not reach sibling share/download hosts,
// even when a response sets a cookie for the whole woozooo.com domain.
type hostCookieJar struct {
	*cookiejar.Jar
}

func (j hostCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		c := *cookie
		c.Domain = ""
		j.Jar.SetCookies(u, []*http.Cookie{&c})
	}
}

func (d *LanZou) getClient(up bool) *resty.Client {
	d.clientOnce.Do(func() {
		jar, _ := cookiejar.New(nil)
		d.client = base.NewRestyClient().SetCookieJar(hostCookieJar{jar})
		if d.IsAccount() || d.IsCookie() {
			r := &http.Request{Header: http.Header{"Cookie": []string{d.Cookie}}}
			d.setAccountCookies(r.Cookies())
		}
	})
	if !up {
		return d.client
	}
	d.uploadOnce.Do(func() {
		d.uploadClient = base.NewRestyClient().SetTimeout(120 * time.Second).
			SetTransport(d.client.GetClient().Transport).
			SetCookieJar(d.client.GetClient().Jar)
	})
	return d.uploadClient
}

// Authentication is shared only with the known account endpoints and the
// configured account API. WAF solutions remain scoped to the issuing host.
func (d *LanZou) setAccountCookies(cookies []*http.Cookie) {
	for _, endpoint := range []string{loginURL, diskURL, d.BaseUrl} {
		u, err := url.Parse(endpoint)
		if err != nil || u.Host == "" {
			continue
		}
		for _, cookie := range cookies {
			if cookie.Name == challengeCookie {
				continue
			}
			c := *cookie
			c.Path = "/"
			c.Secure = u.Scheme == "https"
			d.client.GetClient().Jar.SetCookies(u, []*http.Cookie{&c})
		}
	}
}

func isAcwChallenge(body []byte) bool {
	body = bytes.TrimSpace(body)
	// JSON errors are not challenge pages, even if their message names a cookie.
	return len(body) > 0 && body[0] != '{' && body[0] != '[' &&
		bytes.Contains(body, []byte(challengeCookie))
}

func (d *LanZou) execute(client *resty.Client, endpoint, method string, callback base.ReqCallback, up bool) (*resty.Response, error) {
	for attempt := range 3 {
		req := client.R().SetHeaders(map[string]string{
			"Referer":    "https://pc.woozooo.com",
			"User-Agent": d.UserAgent,
		})
		if callback != nil {
			callback(req)
		}
		if err := req.Context().Err(); err != nil {
			return nil, err
		}
		resp, err := req.Execute(method, endpoint)
		if err != nil {
			return nil, err
		}
		if err := req.Context().Err(); err != nil {
			return nil, err
		}
		if !isAcwChallenge(resp.Body()) {
			return resp, nil
		}
		if up {
			return nil, fmt.Errorf("acw_sc__v2 challenge: cannot safely retry a streaming upload")
		}
		value, err := CalcAcwScV2(string(resp.Body()))
		if err != nil {
			return nil, fmt.Errorf("acw_sc__v2 challenge: %w", err)
		}
		if attempt == 2 {
			break
		}
		// A redirect may have delivered the challenge from a different host.
		u := resp.RawResponse.Request.URL
		client.GetClient().Jar.SetCookies(u, []*http.Cookie{{
			Name:   challengeCookie,
			Value:  value,
			Path:   "/",
			Secure: u.Scheme == "https",
		}})
	}
	return nil, fmt.Errorf("acw_sc__v2 validation failed after 3 attempts")
}
