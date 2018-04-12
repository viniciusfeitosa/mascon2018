package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/user?name=Vinicius Pacheco", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	handler := MessageToUser{}
	handler.ServeHTTP(rec, req)

	if status := rec.Code; status != http.StatusOK {
		t.Errorf("Expected: %d Received: %d", http.StatusOK, status)
	}

	expected := `{"user":"Vinicius Pacheco","message":"Hello World"}`
	if rec.Body.String() != expected {
		t.Errorf("Expected: %s Received: %s", expected, rec.Body.String())
	}
}
