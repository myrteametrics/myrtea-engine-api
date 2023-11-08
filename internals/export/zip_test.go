package export

import (
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"testing"
)

func TestCreatePasswordProtectedZipFile(t *testing.T) {
	file, err := CreatePasswordProtectedZipFile("test.zip", []byte("test"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	expression.AssertNotEqual(t, len(file), 0)
}
