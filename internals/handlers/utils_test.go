package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
)

func TestParseTime(t *testing.T) {
	ti, err := ParseTime("2019-05-05T20:12:00.000+03:00")
	if err != nil {
		t.Error(err)
	}
	_, offset := ti.Zone()
	if ti.Year() != 2019 ||
		ti.Month() != time.May ||
		ti.Day() != 5 ||
		ti.Hour() != 20 ||
		ti.Minute() != 12 ||
		ti.Second() != 0 ||
		ti.Nanosecond() != 0 ||
		offset != 3*60*60 {
		t.Error("invalid date parsing")
		t.Log("ti", ti)
	}
}

func TestParseTimeInvalid(t *testing.T) {
	_, err := ParseTime("2019-05-05T18:99:00.000+02:00")
	if err == nil {
		t.Error("Date should not be parsable")
	}
	_, err = ParseTime("2019-05-05T18:12:00.000")
	if err == nil {
		t.Error("Date should not be parsable")
	}
}

func TestParseTimeEmpty(t *testing.T) {
	_, err := ParseTime("")
	if err == nil {
		t.Error("Parse time should not be possible on empty string")
	}
}

func TestParseSortBy(t *testing.T) {
	var err error
	var sortOptions []models.SortOption
	sortOptions, err = ParseSortBy("asc(field1)", []string{"field1", "field2", "field3"})
	if err != nil {
		t.Error(err)
		t.Log(sortOptions)
	}
	sortOptions, err = ParseSortBy("desc(field1)", []string{"field1", "field2", "field3"})
	if err != nil {
		t.Error(err)
		t.Log(sortOptions)
	}
	sortOptions, err = ParseSortBy("asc(field1),desc(field2),asc(field3)", []string{"field1", "field2", "field3"})
	if err != nil {
		t.Error(err)
		t.Log(sortOptions)
	}
}

func TestParseSortByEmpty(t *testing.T) {
	var err error
	var sortOptions []models.SortOption
	sortOptions, err = ParseSortBy("", []string{"field1", "field2", "field3"})
	if err != nil {
		t.Error(err)
		t.Log(sortOptions)
	}
}

func TestParseSortByInvalid(t *testing.T) {
	var err error
	var sortOptions []models.SortOption
	sortOptions, err = ParseSortBy("asc(aaaa", []string{"field1", "field2", "field3"})
	if err == nil {
		t.Error("sortby expression must be invalid")
		t.Log(sortOptions)
	}
	sortOptions, err = ParseSortBy("asc(field1) desc(field2)", []string{"field1", "field2", "field3"})
	if err == nil {
		t.Error("sortby expression must be invalid")
		t.Log(sortOptions)
	}
	sortOptions, err = ParseSortBy("asc(field1),desc(field2,asc(field3)", []string{"field1", "field2", "field3"})
	if err == nil {
		t.Error("sortby expression must be invalid")
		t.Log(sortOptions)
	}
}

func TestParseSortByInvalidOrder(t *testing.T) {
	var err error
	var sortOptions []models.SortOption
	sortOptions, err = ParseSortBy("other(aaaa)", []string{"field1", "field2", "field3"})
	if err == nil {
		t.Error("sortby expression must be invalid")
		t.Log(sortOptions)
	}
}

func TestParseSortByInvalidField(t *testing.T) {
	var err error
	var sortOptions []models.SortOption
	sortOptions, err = ParseSortBy("asc(other)", []string{"field1", "field2", "field3"})
	if err == nil {
		t.Error("sortby expression must be invalid")
		t.Log(sortOptions)
	}
}

func TestQueryParamToOptionalInt64Array(t *testing.T) {
	r := httptest.NewRequest("GET", "http://test/export", nil)
	query := r.URL.Query()
	query.Add("combineFactIds", "1,2,3,4")
	r.URL.RawQuery = query.Encode()

	expectedResult := []int64{1, 2, 3, 4}
	result, err := QueryParamToOptionalInt64Array(r, "combineFactIds", ",", false, []int64{})

	if err != nil || len(result) != len(expectedResult) {
		t.FailNow()
	}

	for i := 0; i < len(expectedResult); i++ {
		if expectedResult[i] != result[i] {
			t.FailNow()
		}
	}

}

func TestHandleError(t *testing.T) {
	// response writer and request
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	err := errors.New("test error expected")
	handleError(w, r, "test message", err, render.ErrAPIProcessError)

	if err == nil {
		t.Error(err)
	}
	if w.Code != http.StatusInternalServerError {
		t.Error("not a same code Http")
	}
}

func TestGenerateRandomState(t *testing.T) {
	state, err := generateRandomState()

	if err != nil {
		t.Error(err)
	}

	if state == "" {
		t.Error("must not be empty")
	}
}

func TestGenerateEncryptedState(t *testing.T) {
	// encryption key with length 12
	key := []byte("test key")
	encryptedState, err := generateEncryptedState(key)

	if err == nil {
		t.Error(err)
	}

	if encryptedState != "" {
		t.Error("must be empty")
	}

	// encryption key with length 24
	key = []byte("thisis24characterslongs.")
	encryptedState, err = generateEncryptedState(key)

	if err != nil {
		t.Error(err)
	}

	if encryptedState == "" {
		t.Error("must not be empty")
	}

}

func TestVerifyEncryptedState(t *testing.T) {
	// encryption key
	key := []byte("thisis24characterslongs.")

	// Generate an encrypted state
	encryptedState, err := generateEncryptedState(key)

	if err != nil {
		t.Error(err)
	}

	decryptedState, err := verifyEncryptedState(encryptedState, key)

	if err != nil {
		t.Error(err)
	}

	if decryptedState == "" {
		t.Error("must not be empty")
	}

	decryptedState, err = verifyEncryptedState("Fake", key)

	if err == nil {
		t.Error(err)
	}

	if decryptedState != "" {
		t.Error("must be empty")
	}
}
