package handles

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/AlliotTech/openalist/internal/model"
	"github.com/gin-gonic/gin"
)

func TestFsRenameRejectsTraversalName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/fs/rename", strings.NewReader(`{
		"path":"/safe/file.txt",
		"name":"../outside.txt"
	}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", &model.User{
		Role:       model.ADMIN,
		Permission: (1 << 14) - 1,
		BasePath:   "/",
	})

	FsRename(c)

	var response struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, recorder.Body.String())
	}
	if response.Code != http.StatusForbidden {
		t.Fatalf("response code = %d, want %d; body=%s", response.Code, http.StatusForbidden, recorder.Body.String())
	}
}

func TestPlistEscapesBundleIdentifier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	linkName := url.PathEscape("https://example.com/app.ipa") + "/" + url.PathEscape(`safe@"><script>alert(1)</script>`)
	encoded := base64.StdEncoding.EncodeToString([]byte(linkName))
	encoded = strings.NewReplacer("+", "-", "/", "_", "=", ".").Replace(encoded)
	c.Params = []gin.Param{{Key: "link_name", Value: encoded + ".plist"}}

	Plist(c)

	body := recorder.Body.String()
	if strings.Contains(body, "<script>") {
		t.Fatalf("plist contains unescaped script element: %s", body)
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Fatalf("plist does not contain escaped identifier: %s", body)
	}
}
