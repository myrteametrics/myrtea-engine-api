package rule

import (
	"testing"

	"github.com/myrteametrics/myrtea-sdk/v5/ruleeng"
)

func newGroupedSituationReportingRule(parameters map[string]ruleeng.Expression) Rule {
	return Rule{
		Name:        "test-rule",
		Description: "test rule",
		Enabled:     true,
		DefaultRule: ruleeng.DefaultRule{
			Cases: []ruleeng.Case{
				{
					Name:      "case-1",
					Condition: "true",
					Enabled:   true,
					Actions: []ruleeng.ActionDef{
						{
							Name:       `"situation-reporting"`,
							Enabled:    true,
							Parameters: parameters,
						},
					},
				},
			},
		},
	}
}

func TestRuleIsValidGroupBySituationRequiresGroupSubject(t *testing.T) {
	r := newGroupedSituationReportingRule(map[string]ruleeng.Expression{
		"groupBySituation": "true",
	})
	if valid, err := r.IsValid(); valid || err == nil {
		t.Fatalf("expected an error when groupBySituation=true and groupSubject is missing, got valid=%v err=%v", valid, err)
	}
}

func TestRuleIsValidGroupBySituationRejectsAttachments(t *testing.T) {
	r := newGroupedSituationReportingRule(map[string]ruleeng.Expression{
		"groupBySituation":    "true",
		"groupSubject":        `"Alertes regroupées"`,
		"attachmentFileNames": `"export.csv"`,
		"attachmentFactIds":   `"1"`,
	})
	if valid, err := r.IsValid(); valid || err == nil {
		t.Fatalf("expected an error when groupBySituation=true and attachments are configured, got valid=%v err=%v", valid, err)
	}
}

func TestRuleIsValidGroupBySituationWithGroupSubjectAndNoAttachment(t *testing.T) {
	r := newGroupedSituationReportingRule(map[string]ruleeng.Expression{
		"groupBySituation": "true",
		"groupSubject":     `"Alertes regroupées"`,
	})
	if valid, err := r.IsValid(); !valid || err != nil {
		t.Fatalf("expected the rule to be valid, got valid=%v err=%v", valid, err)
	}
}

func TestRuleIsValidWithoutGroupBySituationIsUnaffected(t *testing.T) {
	r := newGroupedSituationReportingRule(map[string]ruleeng.Expression{
		"attachmentFileNames": `"export.csv"`,
		"attachmentFactIds":   `"1"`,
	})
	if valid, err := r.IsValid(); !valid || err != nil {
		t.Fatalf("expected the rule to be valid when groupBySituation is absent, got valid=%v err=%v", valid, err)
	}
}
