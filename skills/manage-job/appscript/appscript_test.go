package appscript

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// mockAppsScript spins up an httptest.Server that mimics the Apps Script web app:
// GET returns JSON directly; POST returns 302 with Location pointing to /redirect,
// which returns the actual response body — same flow as the real deployment.
func mockAppsScript(t *testing.T, getHandler http.HandlerFunc) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/exec", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if getHandler != nil {
				getHandler(w, r)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"success","rows":[]}`))
			}
			return
		}

		// POST: read body, return 302 to /redirect with the body echoed via query param
		body, _ := io.ReadAll(r.Body)
		scheme := "http"
		redirectURL := scheme + "://" + r.Host + "/redirect?payload=" + url.QueryEscape(string(body))
		http.Redirect(w, r, redirectURL, http.StatusFound)
	})

	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := r.URL.Query().Get("payload")
		if payload == "" {
			w.Write([]byte(`{"status":"success"}`))
			return
		}
		w.Write([]byte(payload))
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func TestGet(t *testing.T) {
	server := mockAppsScript(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("industry") != "Tech" {
			t.Errorf("expected industry=Tech query param, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","rows":[{"companyName":"Acme"}]}`))
	})

	app := NewAppScriptWithURL(server.URL + "/exec")
	params := url.Values{}
	params.Set("industry", "Tech")

	result, err := app.Get(params)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp["status"] != "success" {
		t.Errorf("expected status=success, got %v", resp["status"])
	}
}

func TestGet_NoParams(t *testing.T) {
	server := mockAppsScript(t, nil)

	app := NewAppScriptWithURL(server.URL + "/exec")
	result, err := app.Get(nil)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !strings.Contains(result, `"status":"success"`) {
		t.Errorf("unexpected response: %s", result)
	}
}

func TestCreate(t *testing.T) {
	server := mockAppsScript(t, nil)

	app := NewAppScriptWithURL(server.URL + "/exec")
	entry := map[string]interface{}{
		"companyName": "Acme Corp",
		"link":        "https://example.com/job",
		"industry":    "Tech",
		"status":      "Not Started",
	}

	result, err := app.Create(entry)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp["action"] != "create" {
		t.Errorf("expected action=create in echoed payload, got %v", resp["action"])
	}
	if resp["companyName"] != "Acme Corp" {
		t.Errorf("expected companyName=Acme Corp, got %v", resp["companyName"])
	}
}

func TestPatch(t *testing.T) {
	server := mockAppsScript(t, nil)

	app := NewAppScriptWithURL(server.URL + "/exec")
	matchBy := map[string]interface{}{"companyName": "Acme Corp"}
	update := map[string]interface{}{"status": "Interview!"}

	result, err := app.Patch(matchBy, update)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp["action"] != "patch" {
		t.Errorf("expected action=patch, got %v", resp["action"])
	}
	matchByResp, ok := resp["matchBy"].(map[string]interface{})
	if !ok {
		t.Fatalf("matchBy not a map: %T", resp["matchBy"])
	}
	if matchByResp["companyName"] != "Acme Corp" {
		t.Errorf("expected matchBy.companyName=Acme Corp, got %v", matchByResp["companyName"])
	}
	updateResp, ok := resp["update"].(map[string]interface{})
	if !ok {
		t.Fatalf("update not a map: %T", resp["update"])
	}
	if updateResp["status"] != "Interview!" {
		t.Errorf("expected update.status=Interview!, got %v", updateResp["status"])
	}
}

func TestDelete(t *testing.T) {
	server := mockAppsScript(t, nil)

	app := NewAppScriptWithURL(server.URL + "/exec")
	matchBy := map[string]interface{}{"companyName": "Acme Corp"}

	result, err := app.Delete(matchBy)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp["action"] != "delete" {
		t.Errorf("expected action=delete, got %v", resp["action"])
	}
	matchByResp, ok := resp["matchBy"].(map[string]interface{})
	if !ok {
		t.Fatalf("matchBy not a map: %T", resp["matchBy"])
	}
	if matchByResp["companyName"] != "Acme Corp" {
		t.Errorf("expected matchBy.companyName=Acme Corp, got %v", matchByResp["companyName"])
	}
}

func TestPostFollowRedirect_Non302(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	t.Cleanup(server.Close)

	app := NewAppScriptWithURL(server.URL + "/exec")
	_, err := app.Create(map[string]interface{}{"companyName": "Test"})
	if err == nil {
		t.Fatal("expected error for non-302 response, got nil")
	}
	if !strings.Contains(err.Error(), "expected 302") {
		t.Errorf("expected 'expected 302' in error, got: %v", err)
	}
}

func TestPostFollowRedirect_NoLocationHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(server.Close)

	app := NewAppScriptWithURL(server.URL + "/exec")
	_, err := app.Create(map[string]interface{}{"companyName": "Test"})
	if err == nil {
		t.Fatal("expected error for missing Location header, got nil")
	}
	if !strings.Contains(err.Error(), "no redirect Location header") {
		t.Errorf("expected 'no redirect Location header' in error, got: %v", err)
	}
}
