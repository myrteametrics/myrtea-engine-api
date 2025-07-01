package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSON(w, r, map[string]interface{}{"val1": 3, "val2": "test", "val3": map[string]interface{}{"val4": 8}, "val5": []string{"a", "b"}})
	}).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"val1":3,"val2":"test","val3":{"val4":8},"val5":["a","b"]}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestJSONMap(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSON(w, r, map[string]interface{}{
			"test1": map[string]interface{}{"val1": 3, "val2": "test", "val3": map[string]interface{}{"val4": 8}, "val5": []string{"a", "b"}},
			"test2": map[string]interface{}{"val1": 3, "val2": "test", "val3": map[string]interface{}{"val4": 8}, "val5": []string{"a", "b"}},
		})
	}).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"test1":{"val1":3,"val2":"test","val3":{"val4":8},"val5":["a","b"]},"test2":{"val1":3,"val2":"test","val3":{"val4":8},"val5":["a","b"]}}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
