package tests

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
)

// BuildTestHandler build and call a testing handler and returns the resulting ResponseRecorder for validation
func BuildTestHandler(t *testing.T, method string, targetRoute string, body string, handlerRoute string, handler http.HandlerFunc, user interface{}) *httptest.ResponseRecorder {
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewBuffer([]byte(body))
	}
	req, err := http.NewRequest(method, targetRoute, reader)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	switch method {
	case "GET":
		r.Get(handlerRoute, handler)
	case "POST":
		r.Post(handlerRoute, handler)
	case "PUT":
		r.Put(handlerRoute, handler)
	case "DELETE":
		r.Delete(handlerRoute, handler)
	default:
		t.Error("Unknown method", method)
		t.FailNow()
	}

	ctx := context.WithValue(req.Context(), models.ContextKeyUser, user)

	r.ServeHTTP(rr, req.WithContext(ctx))

	return rr
}

// CheckTestHandler checks a ResponseRecorder HTTP status and body
func CheckTestHandler(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int, expectedBody string) {
	if status := rr.Code; status != expectedStatus {
		t.Errorf("handler returned wrong status code: got %v want %v", status, expectedStatus)
	}
	if body := rr.Body.String(); body != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", body, expectedBody)
	}
}
