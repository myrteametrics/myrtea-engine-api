package tasker

import (
	"strings"
	"testing"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
)

func TestBuildSituationReportingTask(t *testing.T) {
	parameters := map[string]interface{}{
		"id":                  "export-1",
		"subject":             "My top CSV export",
		"bodyTemplate":        `{{- range index . "fact-name-1" }} <tr> <td>{{ .key }}</td> <td>{{ .value }}</td> </tr> {{- end }}`,
		"to":                  "to1@gmail.com,to2@gmail.com,to3@gmail.com",
		"attachmentFileNames": "",
		"attachmentFactIds":   "",
		"columns":             "",
		"columnsLabel":        "",
		"smtpUsername":        "",
		"smtpPassword":        "",
		"smtpHost":            "",
		"smtpPort":            "",
	}
	task, err := buildSituationReportingTask(parameters)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(task)

	parameters = map[string]interface{}{
		"id":                  "export-2",
		"subject":             "My top CSV export",
		"bodyTemplate":        `{{- range index . "fact-name-1" }} <tr> <td>{{ .key }}</td> <td>{{ .value }}</td> </tr> {{- end }}`,
		"to":                  "to1@gmail.com,to2@gmail.com,to3@gmail.com",
		"attachmentFileNames": "file1.csv",
		"attachmentFactIds":   "1",
		"columns":             "a,b,c,d.e",
		"columnsLabel":        "Label A,Label B,Label C,Label D.E",
		"smtpUsername":        "",
		"smtpPassword":        "",
		"smtpHost":            "",
		"smtpPort":            "",
	}
	task, err = buildSituationReportingTask(parameters)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(task)
}

func TestBuildMessageBody(t *testing.T) {
	dataItem := map[string]*reader.Item{
		"missing_scheduling": {
			Key:         "",
			KeyAsString: "",
			Aggs:        map[string]*reader.ItemAgg{"doc_count": {Value: 10}},
			Buckets: map[string][]*reader.Item{
				"BySite": {
					{
						Key:         "bucket1",
						KeyAsString: "bucket1",
						Aggs:        map[string]*reader.ItemAgg{"doc_count": {Value: 10}},
					},
					{
						Key:         "bucket2",
						KeyAsString: "bucket2",
						Aggs:        map[string]*reader.ItemAgg{"doc_count": {Value: 10}},
					},
					{
						Key:         "bucket3",
						KeyAsString: "bucket3",
						Aggs:        map[string]*reader.ItemAgg{"doc_count": {Value: 10}},
					},
				},
			},
		},
	}
	templateData := make(map[string]interface{})
	for k, v := range dataItem {
		d, err := v.ToAbstractMap()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		templateData[k] = d
	}
	templateBody := "\"\u003ctable cellspacing='0' cellpadding='1' style='border-collapse: collapse;'\u003e\u003ctr\u003e\u003cth style='border: 1px solid rgb(0,0,0);padding: 5px;font-weight: bold;'\u003eCode Site\u003c/th\u003e\u003cth style='border: 1px solid rgb(0,0,0);padding: 5px;font-weight: bold;'\u003eNb colis manquant\u003c/th\u003e\u003c/tr\u003e{{ range index . 'missing_scheduling' 'buckets' 'BySite' }}\u003ctr\u003e\u003ctd style='border: 1px solid rgb(0,0,0);padding: 5px;'\u003e{{ .key }}\u003c/td\u003e\u003ctd style='border: 1px solid rgb(0,0,0);padding: 5px;'\u003e{{ index . 'aggs' 'doc_count' 'value' }}\u003c/td\u003e\u003c/tr\u003e{{ end }}\u003c/table\u003e\""
	templateBody = strings.ReplaceAll(templateBody, "'", "\"")
	b, err := BuildMessageBody(templateBody, templateData)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(string(b))
}

func TestSituationReportingTask(t *testing.T) {
	t.Skip() // Development test
	parameters := map[string]interface{}{
		"id":                  "export-2",
		"subject":             "My top CSV export",
		"bodyTemplate":        `{{- range index . "fact-name-1" }} <tr> <td>{{ .key }}</td> <td>{{ .value }}</td> </tr> {{- end }}`,
		"to":                  "to1@gmail.com,to2@gmail.com,to3@gmail.com",
		"attachmentFileNames": "file1.csv",
		"attachmentFactIds":   "1",
		"columns":             "a,b,c,d.e",
		"columnsLabel":        "Label A,Label B,Label C,Label D.E",
		"smtpUsername":        "",
		"smtpPassword":        "",
		"smtpHost":            "",
		"smtpPort":            "",
	}
	contextData := ContextData{
		SituationID:        1,
		TS:                 time.Now(),
		TemplateInstanceID: 0,
		RuleID:             1,
		RuleVersion:        1,
		CaseName:           "warn",
	}

	taskExecutionKey := "task-key-1"
	task, err := buildSituationReportingTask(parameters)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	err = task.Perform(taskExecutionKey, contextData)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
