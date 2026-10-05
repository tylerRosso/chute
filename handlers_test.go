// SPDX-License-Identifier: ISC

package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleIndex(t *testing.T) {
	tempDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tempDir, "sample.png"), []byte("data"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := handleIndex(tempDir)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if cc := rr.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("expected Cache-Control: no-cache, got %q", cc)
	}
	etag := rr.Header().Get("ETag")
	if etag == "" || !strings.HasPrefix(etag, `"`) || !strings.HasSuffix(etag, `"`) {
		t.Errorf("expected valid quoted ETag header, got %q", etag)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "chute") {
		t.Errorf("expected page title in HTML output")
	}
	if !strings.Contains(body, "sample.png") {
		t.Errorf("expected rendered file list to include sample.png")
	}
	for _, expected := range []string{`data-mode="light"`, `data-mode="dark"`, `data-mode="auto"`, "chute_theme", "data-theme"} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected %q in HTML output for theme support", expected)
		}
	}
}

func TestHandleIndex_ETagCaching(t *testing.T) {
	tempDir := t.TempDir()
	handler := handleIndex(tempDir)

	// 1. Initial request on empty directory: 200 OK + ETag
	req1, _ := http.NewRequest("GET", "/", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr1.Code)
	}
	etag1 := rr1.Header().Get("ETag")
	if etag1 == "" {
		t.Fatal("expected non-empty ETag on initial response")
	}

	// 2. Subsequent request with matching If-None-Match: 304 Not Modified and empty body
	req2, _ := http.NewRequest("GET", "/", nil)
	req2.Header.Set("If-None-Match", etag1)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusNotModified {
		t.Errorf("expected 304 Not Modified, got %d", rr2.Code)
	}
	if rr2.Body.Len() != 0 {
		t.Errorf("expected empty body on 304 response, got %d bytes", rr2.Body.Len())
	}
	if rr2.Header().Get("ETag") != etag1 {
		t.Errorf("expected matching ETag in 304 header: got %q want %q", rr2.Header().Get("ETag"), etag1)
	}

	// 3. Conditional request with weak ETag prefix (W/)
	reqWeak, _ := http.NewRequest("GET", "/", nil)
	reqWeak.Header.Set("If-None-Match", "W/"+etag1)
	rrWeak := httptest.NewRecorder()
	handler.ServeHTTP(rrWeak, reqWeak)

	if rrWeak.Code != http.StatusNotModified {
		t.Errorf("expected 304 for weak ETag, got %d", rrWeak.Code)
	}

	// 4. Conditional request with wildcard (*)
	reqWildcard, _ := http.NewRequest("GET", "/", nil)
	reqWildcard.Header.Set("If-None-Match", "*")
	rrWildcard := httptest.NewRecorder()
	handler.ServeHTTP(rrWildcard, reqWildcard)

	if rrWildcard.Code != http.StatusNotModified {
		t.Errorf("expected 304 for wildcard If-None-Match, got %d", rrWildcard.Code)
	}

	// 5. Conditional request with comma-separated list including valid tag
	reqMulti, _ := http.NewRequest("GET", "/", nil)
	reqMulti.Header.Set("If-None-Match", `"unrelated-tag", `+etag1)
	rrMulti := httptest.NewRecorder()
	handler.ServeHTTP(rrMulti, reqMulti)

	if rrMulti.Code != http.StatusNotModified {
		t.Errorf("expected 304 for multi-tag list containing match, got %d", rrMulti.Code)
	}

	// 6. Conditional request with non-matching ETag
	reqMismatch, _ := http.NewRequest("GET", "/", nil)
	reqMismatch.Header.Set("If-None-Match", `"different-etag"`)
	rrMismatch := httptest.NewRecorder()
	handler.ServeHTTP(rrMismatch, reqMismatch)

	if rrMismatch.Code != http.StatusOK {
		t.Errorf("expected 200 OK for mismatched ETag, got %d", rrMismatch.Code)
	}

	// 7. Add a file to the directory: ETag must change and cache invalidates
	if err := os.WriteFile(filepath.Join(tempDir, "newfile.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	reqChanged, _ := http.NewRequest("GET", "/", nil)
	reqChanged.Header.Set("If-None-Match", etag1)
	rrChanged := httptest.NewRecorder()
	handler.ServeHTTP(rrChanged, reqChanged)

	if rrChanged.Code != http.StatusOK {
		t.Errorf("expected 200 OK after directory contents changed, got %d", rrChanged.Code)
	}
	etag2 := rrChanged.Header().Get("ETag")
	if etag2 == etag1 {
		t.Errorf("expected ETag to change after file addition, got identical %q", etag2)
	}
	if !strings.Contains(rrChanged.Body.String(), "newfile.txt") {
		t.Errorf("expected body to contain newly added file")
	}

	// 8. HEAD request support with matching ETag
	reqHead, _ := http.NewRequest("HEAD", "/", nil)
	reqHead.Header.Set("If-None-Match", etag2)
	rrHead := httptest.NewRecorder()
	handler.ServeHTTP(rrHead, reqHead)

	if rrHead.Code != http.StatusNotModified {
		t.Errorf("expected 304 for HEAD request with matching ETag, got %d", rrHead.Code)
	}
}

func TestMatchETag(t *testing.T) {
	tests := []struct {
		ifNoneMatch string
		etag        string
		expected    bool
	}{
		{"", `"abc"`, false},
		{"*", `"abc"`, true},
		{`"abc"`, `"abc"`, true},
		{`abc`, `"abc"`, true},
		{`W/"abc"`, `"abc"`, true},
		{`"abc"`, `W/"abc"`, true},
		{`"xyz", "abc"`, `"abc"`, true},
		{`"xyz", W/"abc"`, `"abc"`, true},
		{`"xyz"`, `"abc"`, false},
		{`"abcd"`, `"abc"`, false},
	}

	for _, tt := range tests {
		result := matchETag(tt.ifNoneMatch, tt.etag)
		if result != tt.expected {
			t.Errorf("matchETag(%q, %q) = %v; want %v", tt.ifNoneMatch, tt.etag, result, tt.expected)
		}
	}
}

func TestHandleIndex_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	req, err := http.NewRequest("GET", "/random-path", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := handleIndex(tempDir)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("expected status 404 for invalid route, got %v", status)
	}
}

func TestHandleUpload_SuccessAndSanitization(t *testing.T) {
	tempDir := t.TempDir()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	maliciousFilename := "../../unauthorized_file.txt"
	fileContents := "Secure upload content"

	part, err := writer.CreateFormFile("files", maliciousFilename)
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte(fileContents))
	writer.Close()

	req, err := http.NewRequest("POST", "/upload", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler := handleUpload(tempDir)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedPath := filepath.Join(tempDir, "unauthorized_file.txt")
	savedContent, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("file was not saved at expected sanitized path: %v", err)
	}

	if string(savedContent) != fileContents {
		t.Errorf("content mismatch: got %s want %s", string(savedContent), fileContents)
	}
}

func TestHandleUpload_MethodNotAllowed(t *testing.T) {
	tempDir := t.TempDir()

	req, err := http.NewRequest("GET", "/upload", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := handleUpload(tempDir)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405 Method Not Allowed, got %v", status)
	}
}
