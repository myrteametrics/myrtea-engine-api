package evaluator

import (
	"reflect"
	"strings"
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
	variables, result, err := scanExpression(input, map[string]interface{}{})

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

func TestScanExpressionWithRoundToDecimal(t *testing.T) {
	// Test case based on the issue description:
	// input `"hello world" + roundToDecimal(var1, 10) + " et " + var2`
	// expected result:
	// error: can not evaluate "hello world" + roundToDecimal(var1, 10) + " et " + var2: first argument must be a float64
	// variables: [var1]
	// result: <nil>

	// Input expression
	input := `"hello world" + roundToDecimal(var1, 10) + " et " + var2`

	// Expected results
	expectedVariables := []string{"var1"}
	var expectedResult interface{} = nil

	// Call the function
	variables, result, err := scanExpression(input, map[string]interface{}{})

	// Check for errors - we expect an error in this case
	if err == nil {
		t.Fatalf("scanExpression should have returned an error but didn't")
	}

	// Check that the error message contains the expected text
	expectedErrorText := "first argument must be a float64"
	if !strings.Contains(err.Error(), expectedErrorText) {
		t.Errorf("Expected error to contain %q, got %q", expectedErrorText, err.Error())
	}

	// Check variables
	if !reflect.DeepEqual(variables, expectedVariables) {
		t.Errorf("Expected variables %v, got %v", expectedVariables, variables)
	}

	// Check result
	if result != expectedResult {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}
}
