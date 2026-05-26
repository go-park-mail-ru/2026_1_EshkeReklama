package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSON_EscapesHTML(t *testing.T) {
	rr := httptest.NewRecorder()
	JSON(rr, http.StatusOK, map[string]string{"x": "<script>alert(1)</script>"})
	if strings.Contains(rr.Body.String(), "<script>") {
		t.Fatalf("expected html to be escaped, got %s", rr.Body.String())
	}

	var env struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Data["x"] != "<script>alert(1)</script>" {
		t.Fatalf("expected original string after json decode")
	}
}

func TestDecodeJSON_DisallowUnknownFields(t *testing.T) {
	type payload struct {
		A string `json:"a"`
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":"ok","b":1}`))
	var p payload
	if err := DecodeJSON(req, &p); err == nil {
		t.Fatalf("expected error for unknown field")
	}
}

func TestErrorJSON_EscapesHTML(t *testing.T) {
	rr := httptest.NewRecorder()
	ErrorJSON(rr, http.StatusBadRequest, "<b>bad</b>")
	if strings.Contains(rr.Body.String(), "<b>") {
		t.Fatalf("expected html to be escaped, got %s", rr.Body.String())
	}
}

func TestDecodeJSON_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = nil
	var v map[string]any
	if err := DecodeJSON(req, &v); err == nil {
		t.Fatalf("expected error")
	}
}
