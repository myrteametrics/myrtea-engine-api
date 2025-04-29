package utils

import "testing"

func TestRemoveDuplicates_Int64(t *testing.T) {
	sample := []int64{1, 1, 1, 2, 2, 3, 4}
	expectedResult := []int64{1, 2, 3, 4}
	result := RemoveDuplicates(sample)

	if len(result) != len(expectedResult) {
		t.FailNow()
	}

	for i := 0; i < len(expectedResult); i++ {
		if expectedResult[i] != result[i] {
			t.FailNow()
		}
	}
}

func TestRemoveDuplicates_Int(t *testing.T) {
	sample := []int{1, 1, 1, 2, 2, 3, 4}
	expectedResult := []int{1, 2, 3, 4}
	result := RemoveDuplicates(sample)

	if len(result) != len(expectedResult) {
		t.FailNow()
	}

	for i := 0; i < len(expectedResult); i++ {
		if expectedResult[i] != result[i] {
			t.FailNow()
		}
	}
}

func TestRemoveDuplicates_String(t *testing.T) {
	sample := []string{"a", "a", "a", "b", "b", "c", "d"}
	expectedResult := []string{"a", "b", "c", "d"}
	result := RemoveDuplicates(sample)

	if len(result) != len(expectedResult) {
		t.FailNow()
	}

	for i := 0; i < len(expectedResult); i++ {
		if expectedResult[i] != result[i] {
			t.FailNow()
		}
	}
}
