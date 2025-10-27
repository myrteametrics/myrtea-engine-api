package emailutils

import (
	"bytes"
	"html/template"
	"strings"
)

func BuildMessageBody(templateBody string, templateData map[string]interface{}) ([]byte, error) {
	tmpl, err := template.New("htmlEmail").Funcs(template.FuncMap{
		"split": func(input string, separator string) []string {
			return strings.Split(input, separator)
		},
	}).Parse(templateBody)
	if err != nil {
		return nil, err
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, templateData)
	if err != nil {
		return nil, err
	}

	return body.Bytes(), nil
}

// VerifyMessageBody verifies if the message body template is valid
func VerifyMessageBody(templateBody string) error {
	_, err := template.New("htmlEmail").Funcs(template.FuncMap{
		"split": func(input string, separator string) []string {
			return strings.Split(input, separator)
		},
	}).Parse(templateBody)
	if err != nil {
		return err
	}

	return nil
}
