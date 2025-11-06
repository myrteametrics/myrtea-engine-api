package tasker

import (
	"strings"
	"testing"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/utils/emailutils"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/reader"
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
		"smtpUsername":        "from@gmail.com",
		"smtpPassword":        "",
		"smtpHost":            "testsmtp",
		"smtpPort":            "999",
		"timeout":             "12h",
	}
	task, err := buildSituationReportingTask(parameters)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(task)
	if task.Separator != ',' {
		t.Log("Unexpected Separator")
		t.FailNow()
	}

	parameters = map[string]interface{}{
		"id":                  "export-2",
		"subject":             "My top CSV export",
		"bodyTemplate":        `{{- range index . "fact-name-1" }} <tr> <td>{{ .key }}</td> <td>{{ .value }}</td> </tr> {{- end }}`,
		"to":                  "to1@gmail.com,to2@gmail.com,to3@gmail.com",
		"attachmentFileNames": "file1.csv",
		"attachmentFactIds":   "1",
		"columns":             "a,b,c,d.e",
		"columnsLabel":        "Label A,Label B,Label C,Label D.E",
		"separator":           ";*&",
		"smtpUsername":        "from@gmail.com",
		"smtpPassword":        "",
		"smtpHost":            "testsmtp",
		"smtpPort":            "999",
		"timeout":             "12h",
	}
	task, err = buildSituationReportingTask(parameters)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(task)
	if task.Separator != ';' {
		t.Log("Unexpected Separator")
		t.FailNow()
	}
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

	templateBody := `"
	{{ if le (index . 'missing_scheduling' 'aggs' 'doc_count' 'value') 1.0 }}
	<table cellspacing='0' cellpadding='1' style='border-collapse: collapse'>
		<tr>
			<th style='border: 1px solid rgb(0, 0, 0); padding: 5px; font-weight: bold'>Code Site</th>
			<th style='border: 1px solid rgb(0, 0, 0); padding: 5px; font-weight: bold'>Nb colis manquant</th>
		</tr>
		{{ range index . 'missing_scheduling' 'buckets' 'BySite' }}
		<tr>
			<td style='border: 1px solid rgb(0, 0, 0); padding: 5px'>{{ .key }}</td>
			<td style='border: 1px solid rgb(0, 0, 0); padding: 5px'>{{ index . 'aggs' 'doc_count' 'value' }}</td>
		</tr>
		{{ end }}
	</table>
	{{ else }}
	<div>Aucun site n'est concerné aujourd'hui</div>
	{{ end }}"
	`

	templateBody = strings.ReplaceAll(templateBody, "'", "\"")
	b, err := emailutils.BuildMessageBody(templateBody, templateData)
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
		"smtpUsername":        "a",
		"smtpPassword":        "",
		"smtpHost":            "a",
		"smtpPort":            "1",
		"timeout":             "1h",
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

	// cache[taskExecutionKey] = time.Now().Add(1 * time.Hour)

	err = task.Perform(taskExecutionKey, contextData)
	if err != nil {
		t.Log(err)
		// t.FailNow()
	} else {
		t.Log("1) executed")
	}

	err = task.Perform(taskExecutionKey, contextData)
	if err != nil {
		t.Log(err)
		// t.FailNow()
	} else {
		t.Log("2) executed")
	}
}
