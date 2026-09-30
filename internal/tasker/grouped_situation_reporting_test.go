package tasker

import (
	"strings"
	"testing"
)

func TestBuildSituationReportingTaskGroupBySituation(t *testing.T) {
	baseParameters := map[string]interface{}{
		"id":           "alert-1",
		"subject":      "Alert on instance A",
		"bodyTemplate": "<p>content</p>",
		"to":           "to1@gmail.com",
		"timeout":      "1h",
	}

	// groupBySituation defaults to false, groupSubject is not required
	task, err := buildSituationReportingTask(baseParameters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.GroupBySituation {
		t.Fatalf("expected GroupBySituation to default to false")
	}

	// groupBySituation must be a real bool (evaluated expression), like isNotification on create-issue
	invalidType := map[string]interface{}{}
	for k, v := range baseParameters {
		invalidType[k] = v
	}
	invalidType["groupBySituation"] = "true" // string, not bool: must be rejected
	if _, err := buildSituationReportingTask(invalidType); err == nil {
		t.Fatalf("expected error when groupBySituation is a string instead of a bool")
	}

	// groupBySituation=true without groupSubject must fail
	grouped := map[string]interface{}{}
	for k, v := range baseParameters {
		grouped[k] = v
	}
	grouped["groupBySituation"] = true
	if _, err := buildSituationReportingTask(grouped); err == nil {
		t.Fatalf("expected error when groupBySituation=true and groupSubject is missing")
	}

	// groupBySituation=true with groupSubject must succeed
	grouped["groupSubject"] = "Alertes regroupées - Zone Nord"
	task, err = buildSituationReportingTask(grouped)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !task.GroupBySituation {
		t.Fatalf("expected GroupBySituation to be true")
	}
	if task.GroupSubject != "Alertes regroupées - Zone Nord" {
		t.Fatalf("unexpected GroupSubject: %s", task.GroupSubject)
	}
}

func TestBuildSituationReportingTaskGroupBySituationDropsAttachments(t *testing.T) {
	parameters := map[string]interface{}{
		"id":                  "alert-1",
		"subject":             "Alert",
		"bodyTemplate":        "<p>content</p>",
		"to":                  "to1@gmail.com",
		"timeout":             "1h",
		"attachmentFileNames": "export.csv",
		"attachmentFactIds":   "1",
		"groupBySituation":    true,
		"groupSubject":        "Alertes regroupées",
	}

	task, err := buildSituationReportingTask(parameters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(task.AttachmentFileNames) != 0 || len(task.AttachmentFactIDs) != 0 {
		t.Fatalf("expected attachments to be dropped when groupBySituation is true, got names=%v ids=%v",
			task.AttachmentFileNames, task.AttachmentFactIDs)
	}
}

func TestCollectGroupedSituationReporting(t *testing.T) {
	pending := make(map[string]*groupedReportAcc)

	// Two instances of the same situation, same rule, same task id: same group.
	task1 := SituationReportingTask{
		ID:               "alert-1",
		Subject:          "Instance A is critical",
		BodyTemplate:     "<p>value: {{.value}}</p>",
		To:               []string{"to1@gmail.com"},
		Cc:               []string{"cc1@gmail.com"},
		Timeout:          "1h",
		GroupBySituation: true,
		GroupSubject:     "Alertes regroupées - Zone Nord",
	}
	context1 := ContextData{
		SituationID:                 10,
		RuleID:                      100,
		TemplateInstanceID:          1,
		HistorySituationFlattenData: map[string]interface{}{"value": "A"},
	}

	task2 := task1
	task2.Subject = "Instance B is critical"
	task2.To = []string{"to2@gmail.com"} // distinct recipient: must be unioned, not lost
	context2 := context1
	context2.TemplateInstanceID = 2
	context2.HistorySituationFlattenData = map[string]interface{}{"value": "B"}

	if err := collectGroupedSituationReporting(pending, task1, context1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := collectGroupedSituationReporting(pending, task2, context2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pending) != 1 {
		t.Fatalf("expected exactly 1 group, got %d", len(pending))
	}

	var acc *groupedReportAcc
	for _, a := range pending {
		acc = a
	}

	if len(acc.members) != 2 {
		t.Fatalf("expected 2 members in the group, got %d", len(acc.members))
	}
	if len(acc.to) != 2 || acc.to[0] != "to1@gmail.com" || acc.to[1] != "to2@gmail.com" {
		t.Fatalf("expected recipients to be unioned, got %v", acc.to)
	}
	if len(acc.cc) != 1 || acc.cc[0] != "cc1@gmail.com" {
		t.Fatalf("expected cc to be preserved, got %v", acc.cc)
	}

	body := buildGroupedReportBody(acc.members)
	if !strings.Contains(body, "regroupe 2 alertes") {
		t.Fatalf("expected the body to mention the group size, got: %s", body)
	}
	if !strings.Contains(body, "Instance A is critical") || !strings.Contains(body, "value: A") {
		t.Fatalf("expected member A's subject and content in the body, got: %s", body)
	}
	if !strings.Contains(body, "Instance B is critical") || !strings.Contains(body, "value: B") {
		t.Fatalf("expected member B's subject and content in the body, got: %s", body)
	}
}

func TestBuildGroupTaskKeyIgnoresInstance(t *testing.T) {
	task := SituationReportingTask{ID: "alert-1"}
	context1 := ContextData{SituationID: 10, RuleID: 100, TemplateInstanceID: 1}
	context2 := ContextData{SituationID: 10, RuleID: 100, TemplateInstanceID: 2}

	if buildGroupTaskKey(context1, task) != buildGroupTaskKey(context2, task) {
		t.Fatalf("expected the group key to be identical across instances of the same situation")
	}
}

func TestBuildSituationReportingKeyIsPerInstance(t *testing.T) {
	task := SituationReportingTask{ID: "alert-1"}
	context1 := ContextData{SituationID: 10, RuleID: 100, TemplateInstanceID: 1}
	context2 := ContextData{SituationID: 10, RuleID: 100, TemplateInstanceID: 2}

	if buildSituationReportingKey(context1, task) == buildSituationReportingKey(context2, task) {
		t.Fatalf("expected the per-instance key to differ across instances of the same situation")
	}
}
