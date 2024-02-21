package main

import (
	"go-app/polytopia"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPolytopiaGetGood(t *testing.T) {
    req, err := http.NewRequest("GET", "/", nil)
    if err != nil {
        t.Fatal(err)
    }
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(polytopia.PolytopiaHandler)
    handler.ServeHTTP(rr, req)
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v",
        status, http.StatusOK)
    }
}

func TestPolytopiaBadMethod(t *testing.T) {
    req, err := http.NewRequest("PUT", "/", nil)
    if err != nil {
        t.Fatal(err)
    }
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(polytopia.PolytopiaHandler)
    handler.ServeHTTP(rr, req)
    if status := rr.Code; status != http.StatusMethodNotAllowed {
        t.Errorf("handler returned wrong status code: got %v want %v",
        status, http.StatusMethodNotAllowed)
    }
}
