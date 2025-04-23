package situation

import (
	"reflect"
	"testing"
)

func TestGetTranslateValue(t *testing.T) {

	if shouldParseGlobalVariables() != true {
		t.Error("Expected true for empty input, got true")
	}

	if shouldParseGlobalVariables(false) != false {
		t.Error("Expected false for input of false, got true")
	}

	if shouldParseGlobalVariables(true) != true {
		t.Error("Expected true for input of true, got false")
	}

	if shouldParseGlobalVariables(true, false) != true {
		t.Error("Expected true for multiple input values, but function should only consider the first")
	}
}

func TestReplaceKeysWithValues(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		vars     map[string]interface{}
		expected map[string]string
	}{
		{
			name: "string + key unique + String",
			input: map[string]string{
				"lien_adel": "\"https://\"+ keyUnique +\"/url\"",
			},
			vars: map[string]interface{}{
				"keyUnique": "url.exemple.eu",
			},
			expected: map[string]string{
				"lien_adel": "https://url.exemple.eu/url",
			},
		},
		{
			name: "key unique + String",
			input: map[string]string{
				"lien_adel": "keyUnique +\"/url\"",
			},
			vars: map[string]interface{}{
				"keyUnique": "https://url.exemple.eu",
			},
			expected: map[string]string{
				"lien_adel": "https://url.exemple.eu/url",
			},
		},
		{
			name: "String + key unique",
			input: map[string]string{
				"lien_adel": "\"https://\"+keyUnique",
			},
			vars: map[string]interface{}{
				"keyUnique": "url.exemple.eu/url",
			},
			expected: map[string]string{
				"lien_adel": "https://url.exemple.eu/url",
			},
		},
		{
			name: "key unique",
			input: map[string]string{
				"lien_adel": "keyUnique",
			},
			vars: map[string]interface{}{
				"keyUnique": "https://url.exemple.eu/url",
			},
			expected: map[string]string{
				"lien_adel": "https://url.exemple.eu/url",
			},
		},
		{
			name: "plusieur key unique",
			input: map[string]string{
				"lien_adel": "\"https://\"+key1+\"url\"+key2+\"domain\"+key3",
			},
			vars: map[string]interface{}{
				"key1": "firstPart",
				"key2": "secondPart",
				"key3": "thirdPart",
			},
			expected: map[string]string{
				"lien_adel": "https://firstParturlsecondPartdomainthirdPart",
			},
		},
		{
			name: "no variable unique",
			input: map[string]string{
				"lien_adel": "keyUnique yyy",
			},
			vars: map[string]interface{}{
				"key1": "firstPart",
				"key2": "secondPart",
				"key3": "thirdPart",
			},
			expected: map[string]string{
				"lien_adel": "keyUnique yyy",
			},
		},
		{
			name: "boolean",
			input: map[string]string{
				"lien_adel": "false",
			},
			vars: map[string]interface{}{
				"key1": "firstPart",
				"key2": "secondPart",
				"key3": "thirdPart",
			},
			expected: map[string]string{
				"lien_adel": "false",
			},
		},
		{
			name: "lien",
			input: map[string]string{
				"lien_adel": "https://wiki.alturing.eu/bin/view/Main/H24/FLUX/FLUX%20CLP/%F0%9F%93%82%20MYRTEA/",
			},
			vars: map[string]interface{}{
				"key1": "firstPart",
				"key2": "secondPart",
				"key3": "thirdPart",
			},
			expected: map[string]string{
				"lien_adel": "https://wiki.alturing.eu/bin/view/Main/H24/FLUX/FLUX%20CLP/%F0%9F%93%82%20MYRTEA/",
			},
		},
		{
			name: "nombre",
			input: map[string]string{
				"lien_adel": "\"123456789\"",
			},
			vars: map[string]interface{}{
				"key1": "firstPart",
				"key2": "secondPart",
				"key3": "thirdPart",
			},
			expected: map[string]string{
				"lien_adel": "123456789",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ReplaceKeysWithValues(tt.input, tt.vars)
			if !reflect.DeepEqual(tt.input, tt.expected) {
				t.Errorf("replaceKeysWithValues() = %v, want %v", tt.input, tt)
			}
		})
	}
}
