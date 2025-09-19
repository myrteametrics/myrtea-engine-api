package evaluator

import (
	"reflect"
	"testing"
)

func TestScanExpression(t *testing.T) {
	// Test case based on the comment:
	// input `"hello world " + var1 + " et " + var2`
	// expected result:
	// variables: [var1 var2]
	// expression result: hello world {{ var1 }} et {{ var2 }}

	// Input expression
	input := `"hello world " + var1 + " et " + var2`

	// Expected results
	expectedVariables := []string{"var1", "var2"}
	expectedResult := "hello world {{ var1 }} et {{ var2 }}"

	// Call the function
	variables, result, err := ScanExpression(input, map[string]interface{}{}, false)

	// Check for errors
	if err != nil {
		t.Fatalf("scanExpression returned an error: %v", err)
	}

	// Check variables
	if !reflect.DeepEqual(variables, expectedVariables) {
		t.Errorf("Expected variables %v, got %v", expectedVariables, variables)
	}

	// Check result
	if result != expectedResult {
		t.Errorf("Expected result %q, got %q", expectedResult, result)
	}
}
