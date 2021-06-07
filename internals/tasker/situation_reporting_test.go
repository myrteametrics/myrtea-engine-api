package tasker

import (
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
		"fact-1": {
			Key:         "",
			KeyAsString: "",
			Aggs:        map[string]*reader.ItemAgg{"total": {Value: 10}},
			Buckets: map[string][]*reader.Item{
				"ByCustomBucket": {
					{
						Key:         "bucket1",
						KeyAsString: "bucket1",
						Aggs:        map[string]*reader.ItemAgg{"total": {Value: 10}},
					},
					{
						Key:         "bucket2",
						KeyAsString: "bucket2",
						Aggs:        map[string]*reader.ItemAgg{"total": {Value: 10}},
					},
					{
						Key:         "bucket3",
						KeyAsString: "bucket3",
						Aggs:        map[string]*reader.ItemAgg{"total": {Value: 10}},
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
	b, err := BuildMessageBody(`
		{{ range index . "fact-1" "buckets" "ByCustomBucket" }} 
			{{ .key }} -> {{ index . "aggs" "total" "value" }}			
		{{ end }}
	`, templateData)
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
