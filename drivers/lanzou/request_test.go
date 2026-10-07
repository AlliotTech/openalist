package lanzou

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/AlliotTech/openalist/internal/conf"
	"github.com/go-resty/resty/v2"
)

const authChallengeHTML = `<html><script>var arg1='0CCD57BCF5ACEE5B2DF878457E6DCF03D1AF6954';document.cookie='acw_sc__v2';</script></html>`
const authChallengeValue = "6ac5cc695f752682786f5cdfd73da6bd024d15ba"

// The real http.Client still processes cookies and redirects; only its network
// transport is replaced, so the fixed account endpoints never reach a network.
type lanzouTestTransport func(*http.Request) (*http.Response, error)

func (f lanzouTestTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newAuthTestDriver(t *testing.T, handler http.HandlerFunc) *LanZou {
	t.Helper()
	if conf.Conf == nil {
		conf.Conf = conf.DefaultConfig()
		t.Cleanup(func() { conf.Conf = nil })
	}
	d := &LanZou{Addition: Addition{
		Type:      "account",
		Account:   "account + name",
		Password:  "password & value",
		BaseUrl:   "https://pc.woozooo.com",
		ShareUrl:  "https://share.woozooo.com",
		UserAgent: "Lanzou authentication test",
	}}
	d.getClient(false).SetTransport(lanzouTestTransport(func(r *http.Request) (*http.Response, error) {
		if r.Body != nil {
			defer r.Body.Close()
		}
		recorder := httptest.NewRecorder()
		handler(recorder, r)
		resp := recorder.Result()
		resp.Request = r
		return resp, nil
	}))
	return d
}

func requireAuthCookie(t *testing.T, r *http.Request, name, value string) {
	t.Helper()
	cookie, err := r.Cookie(name)
	if err != nil || cookie.Value != value {
		t.Fatalf("cookie %s = %v, %v; want %q", name, cookie, err, value)
	}
	count := 0
	for _, cookie := range r.Cookies() {
		if cookie.Name == name {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("cookie %s occurs %d times", name, count)
	}
}

func requireNoAuthCookie(t *testing.T, r *http.Request, name string) {
	t.Helper()
	if cookie, err := r.Cookie(name); err != http.ErrNoCookie {
		t.Fatalf("unexpected cookie %s: %v, %v", name, cookie, err)
	}
}

func TestLoginChallengePreservesFormAndCookies(t *testing.T) {
	requests := 0
	d := newAuthTestDriver(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodPost || r.URL.String() != loginURL {
			t.Fatalf("unexpected login request: %s %s", r.Method, r.URL)
		}
		if r.UserAgent() != "Lanzou authentication test" {
			t.Fatalf("User-Agent = %q", r.UserAgent())
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		want := url.Values{
			"task":         {"3"},
			"uid":          {"account + name"},
			"pwd":          {"password & value"},
			"setSessionId": {""},
			"setSig":       {""},
			"setScene":     {""},
			"setTocen":     {""},
			"formhash":     {""},
		}
		if !reflect.DeepEqual(r.PostForm, want) {
			t.Fatalf("login form = %v, want %v", r.PostForm, want)
		}
		switch requests {
		case 1:
			requireNoAuthCookie(t, r, challengeCookie)
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "intermediate", Domain: ".woozooo.com", Path: "/"})
			fmt.Fprint(w, authChallengeHTML)
		case 2:
			requireAuthCookie(t, r, challengeCookie, authChallengeValue)
			requireAuthCookie(t, r, "session", "intermediate")
			http.SetCookie(w, &http.Cookie{Name: "ylogin", Value: "authenticated", Domain: ".woozooo.com", Path: "/"})
			fmt.Fprint(w, `{"zt":1}`)
		default:
			t.Fatalf("unexpected login attempt %d", requests)
		}
	})
	cookies, err := d.Login(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Fatalf("login attempts = %d, want 2", requests)
	}
	stored := &http.Request{Header: http.Header{"Cookie": {d.Cookie}}}
	requireAuthCookie(t, stored, "session", "intermediate")
	requireAuthCookie(t, stored, "ylogin", "authenticated")
	requireNoAuthCookie(t, stored, challengeCookie)
	if len(cookies) != 2 {
		t.Fatalf("returned cookies = %v", cookies)
	}
}

func TestLoginResponses(t *testing.T) {
	for _, tc := range []struct {
		name     string
		body     string
		attempts int
		errText  string
	}{
		{name: "plain success", body: `{"zt":1}`, attempts: 1},
		{name: "password error", body: `{"zt":0,"inf":"wrong password"}`, attempts: 1, errText: "wrong password"},
		{name: "JSON mentioning cookie", body: `{"zt":0,"inf":"acw_sc__v2"}`, attempts: 1, errText: "login err"},
		{name: "persistent challenge", body: authChallengeHTML, attempts: 3, errText: "after 3 attempts"},
		{name: "missing argument", body: `<html>acw_sc__v2</html>`, attempts: 1, errText: "acw_sc__v2"},
		{name: "invalid argument", body: `<script>var arg1='123';document.cookie='acw_sc__v2';</script>`, attempts: 1, errText: "acw_sc__v2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := 0
			d := newAuthTestDriver(t, func(w http.ResponseWriter, r *http.Request) {
				requests++
				if requests > 1 {
					requireAuthCookie(t, r, challengeCookie, authChallengeValue)
				}
				fmt.Fprint(w, tc.body)
			})
			_, err := d.Login(context.Background())
			if tc.errText == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.errText) {
				t.Fatalf("error = %v, want %q", err, tc.errText)
			}
			if requests != tc.attempts {
				t.Fatalf("requests = %d, want %d", requests, tc.attempts)
			}
		})
	}
}

func TestInitHandlesAccountPageChallenge(t *testing.T) {
	ctx := context.Background()
	loginRequests, diskRequests := 0, 0
	d := newAuthTestDriver(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mlogin.php" {
			loginRequests++
			if loginRequests == 1 {
				fmt.Fprint(w, authChallengeHTML)
				return
			}
			requireAuthCookie(t, r, challengeCookie, authChallengeValue)
			http.SetCookie(w, &http.Cookie{Name: "ylogin", Value: "account", Path: "/"})
			fmt.Fprint(w, `{"zt":1}`)
			return
		}
		if r.URL.Host != "pc.woozooo.com" || r.URL.Path != "/mydisk.php" || r.Method != http.MethodGet {
			t.Fatalf("unexpected account page request: %s %s", r.Method, r.URL)
		}
		if got := r.URL.Query(); !reflect.DeepEqual(got, url.Values{"item": {"files"}, "action": {"index"}}) {
			t.Fatalf("account page query = %v", got)
		}
		diskRequests++
		if diskRequests == 1 {
			requireAuthCookie(t, r, "ylogin", "account")
			requireNoAuthCookie(t, r, challengeCookie)
			http.SetCookie(w, &http.Cookie{Name: "ylogin", Value: "refreshed", Path: "/"})
			fmt.Fprint(w, authChallengeHTML)
			return
		}
		requireAuthCookie(t, r, "ylogin", "refreshed")
		requireAuthCookie(t, r, challengeCookie, authChallengeValue)
		fmt.Fprint(w, `<a href="/doupload.php?uid=42&task=5"></a><script>data: {'vei':'disk-token'}</script>`)
	})
	if err := d.Init(ctx); err != nil {
		t.Fatal(err)
	}
	if d.uid != "42" || d.vei != "disk-token" {
		t.Fatalf("account initialization: uid=%q vei=%q", d.uid, d.vei)
	}
	if loginRequests != 2 || diskRequests != 2 {
		t.Fatalf("attempts: login=%d disk=%d, want 2 each", loginRequests, diskRequests)
	}
}

func TestInitCancellation(t *testing.T) {
	for _, stage := range []string{"before login", "login challenge", "account page challenge"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			requests := 0
			d := newAuthTestDriver(t, func(w http.ResponseWriter, r *http.Request) {
				requests++
				if stage == "account page challenge" && r.URL.Path == "/mlogin.php" {
					fmt.Fprint(w, `{"zt":1}`)
					return
				}
				cancel()
				fmt.Fprint(w, authChallengeHTML)
			})
			want := 1
			if stage == "before login" {
				cancel()
				want = 0
			} else if stage == "account page challenge" {
				want = 2
			}
			if err := d.Init(ctx); !errors.Is(err, context.Canceled) {
				t.Fatalf("Init error = %v, want cancellation", err)
			}
			if requests != want {
				t.Fatalf("requests = %d, want %d", requests, want)
			}
		})
	}
}

func TestLoginDoesNotFollowRedirects(t *testing.T) {
	requests := 0
	d := newAuthTestDriver(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.String() != loginURL {
			t.Fatalf("login followed redirect to %s", r.URL)
		}
		http.Redirect(w, r, "https://share.woozooo.com/target", http.StatusFound)
	})
	if _, err := d.Login(context.Background()); err == nil {
		t.Fatal("redirected login unexpectedly succeeded")
	}
	if requests == 0 {
		t.Fatal("login did not reach the transport")
	}
}

func TestSessionCookiesAreAccountAndHostIsolated(t *testing.T) {
	for _, account := range []string{"first", "second"} {
		loginRequests := 0
		d := newAuthTestDriver(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/mlogin.php" {
				loginRequests++
				if loginRequests == 1 {
					if got := r.Header.Get("Cookie"); got != "" {
						t.Fatalf("new account %s inherited cookies: %s", account, got)
					}
					http.SetCookie(w, &http.Cookie{Name: "ylogin", Value: account, Domain: ".woozooo.com", Path: "/"})
					fmt.Fprint(w, authChallengeHTML)
					return
				}
				requireAuthCookie(t, r, "ylogin", account)
				requireAuthCookie(t, r, challengeCookie, authChallengeValue)
				fmt.Fprint(w, `{"zt":1}`)
				return
			}
			switch r.URL.Host {
			case "up.woozooo.com":
				requireAuthCookie(t, r, "ylogin", account)
				requireAuthCookie(t, r, challengeCookie, authChallengeValue)
			case "pc.woozooo.com":
				requireAuthCookie(t, r, "ylogin", account)
				requireNoAuthCookie(t, r, challengeCookie)
				if r.URL.Path == "/redirect" {
					http.Redirect(w, r, "https://sub.pc.woozooo.com/target", http.StatusFound)
					return
				}
			default:
				if got := r.Header.Get("Cookie"); got != "" {
					t.Fatalf("cookies leaked to %s: %s", r.URL, got)
				}
			}
			fmt.Fprint(w, "ok")
		})
		d.Account = account
		if _, err := d.Login(context.Background()); err != nil {
			t.Fatal(err)
		}
		for _, endpoint := range []string{
			"https://up.woozooo.com/check",
			"https://pc.woozooo.com/check",
			"https://pc.woozooo.com/redirect",
			"https://share.woozooo.com/check",
			"https://other.example/check",
			"https://download.lanzou.example/check",
		} {
			if _, err := d.get(endpoint, nil); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestShareChallengeUsesGenericRequest(t *testing.T) {
	requests := 0
	d := newAuthTestDriver(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			fmt.Fprint(w, authChallengeHTML)
			return
		}
		requireAuthCookie(t, r, challengeCookie, authChallengeValue)
		fmt.Fprint(w, "文件不存在")
	})
	if _, err := d.getShareUrlHtml("share-id"); !errors.Is(err, ErrFileNotExist) {
		t.Fatalf("share error = %v, want ErrFileNotExist", err)
	}
	if requests != 2 {
		t.Fatalf("share requests = %d, want 2", requests)
	}
}

type uploadOnlyReader struct {
	io.Reader
}

func TestStreamingUploadChallengeIsNotReplayed(t *testing.T) {
	for _, challenged := range []bool{false, true} {
		t.Run(fmt.Sprintf("challenge=%t", challenged), func(t *testing.T) {
			requests, configured := 0, 0
			d := newAuthTestDriver(t, func(w http.ResponseWriter, r *http.Request) {
				requests++
				if err := r.ParseMultipartForm(1024); err != nil {
					t.Fatal(err)
				}
				defer r.MultipartForm.RemoveAll()
				file, header, err := r.FormFile("upload_file")
				if err != nil {
					t.Fatal(err)
				}
				defer file.Close()
				body, err := io.ReadAll(file)
				if err != nil {
					t.Fatal(err)
				}
				if string(body) != "streamed upload" || header.Filename != "test.txt" || r.FormValue("task") != "1" {
					t.Fatalf("unexpected multipart upload: name=%q task=%q body=%q", header.Filename, r.FormValue("task"), body)
				}
				if challenged {
					fmt.Fprint(w, authChallengeHTML)
				} else {
					fmt.Fprint(w, `{"zt":1}`)
				}
			})
			reader := uploadOnlyReader{Reader: strings.NewReader("streamed upload")}
			_, err := d._post(d.BaseUrl+"/html5up.php", func(req *resty.Request) {
				configured++
				req.SetFileReader("upload_file", "test.txt", reader).SetFormData(map[string]string{"task": "1"})
			}, nil, true)
			if challenged {
				if err == nil || !strings.Contains(err.Error(), "cannot safely retry a streaming upload") {
					t.Fatalf("upload error = %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if requests != 1 || configured != 1 {
				t.Fatalf("stream replayed: requests=%d configured=%d", requests, configured)
			}
		})
	}
}

func TestCookieInitializationAndConfiguredAccountHost(t *testing.T) {
	requests := 0
	transportDriver := newAuthTestDriver(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		switch r.URL.Host {
		case "pc.woozooo.com":
			requireAuthCookie(t, r, "ylogin", "configured")
			if requests == 1 {
				fmt.Fprint(w, authChallengeHTML)
				return
			}
			requireAuthCookie(t, r, challengeCookie, authChallengeValue)
			fmt.Fprint(w, `uid=17&task=5; data: {'vei':'cookie-token'}`)
		case "account.example":
			requireAuthCookie(t, r, "ylogin", "configured")
			requireNoAuthCookie(t, r, challengeCookie)
			fmt.Fprint(w, "ok")
		default:
			if got := r.Header.Get("Cookie"); got != "" {
				t.Fatalf("configured cookie leaked to %s: %s", r.URL, got)
			}
			fmt.Fprint(w, "ok")
		}
	})
	d := &LanZou{Addition: Addition{
		Type:    "cookie",
		Cookie:  "ylogin=configured",
		BaseUrl: "https://account.example",
	}}
	d.getClient(false).SetTransport(transportDriver.getClient(false).GetClient().Transport)
	if err := d.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	if requests != 2 || d.uid != "17" || d.vei != "cookie-token" {
		t.Fatalf("cookie initialization: requests=%d uid=%q vei=%q", requests, d.uid, d.vei)
	}
	for _, endpoint := range []string{d.BaseUrl, "https://share.example", "https://sub.account.example"} {
		if _, err := d.get(endpoint, nil); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAccountPostChallengePreservesParameters(t *testing.T) {
	requests := 0
	d := newAuthTestDriver(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodPost || r.URL.Path != "/doupload.php" {
			t.Fatalf("unexpected account request: %s %s", r.Method, r.URL)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(r.URL.Query(), url.Values{"uid": {"42"}, "vei": {"token"}}) ||
			!reflect.DeepEqual(r.PostForm, url.Values{"task": {"47"}, "folder_id": {"-1"}}) {
			t.Fatalf("account parameters changed: query=%v form=%v", r.URL.Query(), r.PostForm)
		}
		if requests == 1 {
			fmt.Fprint(w, authChallengeHTML)
			return
		}
		requireAuthCookie(t, r, challengeCookie, authChallengeValue)
		fmt.Fprint(w, `{"zt":1,"text":"done"}`)
	})
	d.uid, d.vei = "42", "token"
	var result RespText[string]
	_, err := d.doupload(func(req *resty.Request) {
		req.SetFormData(map[string]string{"task": "47", "folder_id": "-1"})
	}, &result)
	if err != nil || result.Text != "done" || requests != 2 {
		t.Fatalf("account response: text=%q requests=%d err=%v", result.Text, requests, err)
	}
}

func TestConcurrentHostChallenges(t *testing.T) {
	var mu sync.Mutex
	counts := make(map[string]int)
	d := newAuthTestDriver(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		counts[r.URL.Host]++
		attempt := counts[r.URL.Host]
		mu.Unlock()
		if attempt == 1 {
			if got := r.Header.Get("Cookie"); got != "" {
				t.Errorf("new host %s inherited cookies: %s", r.URL.Host, got)
			}
			fmt.Fprint(w, authChallengeHTML)
			return
		}
		cookie, err := r.Cookie(challengeCookie)
		if err != nil || cookie.Value != authChallengeValue {
			t.Errorf("host %s challenge cookie = %v, %v", r.URL.Host, cookie, err)
		}
		fmt.Fprint(w, "ok")
	})
	var wg sync.WaitGroup
	for index := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body, err := d.get(fmt.Sprintf("https://share%d.woozooo.com/file", index), nil)
			if err != nil || string(body) != "ok" {
				t.Errorf("host %d: body=%q err=%v", index, body, err)
			}
		}()
	}
	wg.Wait()
	for host, count := range counts {
		if count != 2 {
			t.Errorf("host %s made %d requests, want 2", host, count)
		}
	}
}

func TestRedirectChallengeStaysOnResponseHost(t *testing.T) {
	originRequests, shareRequests := 0, 0
	d := newAuthTestDriver(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Host {
		case "pc.woozooo.com":
			originRequests++
			requireNoAuthCookie(t, r, challengeCookie)
			http.Redirect(w, r, "https://share.woozooo.com/file", http.StatusFound)
		case "share.woozooo.com":
			shareRequests++
			if shareRequests == 1 {
				fmt.Fprint(w, authChallengeHTML)
				return
			}
			requireAuthCookie(t, r, challengeCookie, authChallengeValue)
			fmt.Fprint(w, "share page")
		default:
			t.Fatalf("unexpected redirect host: %s", r.URL.Host)
		}
	})
	body, err := d.get("https://pc.woozooo.com/share", nil)
	if err != nil || string(body) != "share page" || originRequests != 2 || shareRequests != 2 {
		t.Fatalf("redirect challenge: body=%q origin=%d share=%d err=%v", body, originRequests, shareRequests, err)
	}
}
