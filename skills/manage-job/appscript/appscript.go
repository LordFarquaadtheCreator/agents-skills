package appscript

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// AppScript encapsulates communication with the deployed Apps Script web app.
// Commands call methods on this struct — they don't deal with URLs,
// redirects, or payload construction directly.
type AppScript struct {
	url string
}

func NewAppScript() *AppScript {
	return &AppScript{url: loadScriptURL()}
}

func NewAppScriptWithURL(u string) *AppScript {
	return &AppScript{url: u}
}

// postFollowRedirect handles Apps Script's 302 redirect on POST.
// Go's default client converts POST to GET on redirect, dropping the body.
// We capture the Location header, then GET it to retrieve the actual response.
func (a *AppScript) postFollowRedirect(body []byte) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("POST", a.url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 302 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected 302, got %d: %s", resp.StatusCode, string(respBody))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no redirect Location header")
	}

	redirectResp, err := http.Get(location)
	if err != nil {
		return "", err
	}
	defer redirectResp.Body.Close()

	respBody, err := io.ReadAll(redirectResp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

func (a *AppScript) Get(params url.Values) (string, error) {
	target := a.url
	if len(params) > 0 {
		target = target + "?" + params.Encode()
	}

	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (a *AppScript) Create(entry map[string]interface{}) (string, error) {
	entry["action"] = "create"
	body, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return a.postFollowRedirect(body)
}

func (a *AppScript) Patch(matchBy, update map[string]interface{}) (string, error) {
	payload := map[string]interface{}{
		"action":  "patch",
		"matchBy": matchBy,
		"update":  update,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return a.postFollowRedirect(body)
}

func (a *AppScript) Delete(matchBy map[string]interface{}) (string, error) {
	payload := map[string]interface{}{
		"action":  "delete",
		"matchBy": matchBy,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return a.postFollowRedirect(body)
}
