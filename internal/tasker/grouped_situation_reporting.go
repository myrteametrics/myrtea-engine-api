package tasker

import (
	"fmt"
	"html"
	"strings"
	"time"

	email2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/email"
	"go.uber.org/zap"
)

// groupedReportMember is one instance's rendered report, waiting to be merged into a single email.
type groupedReportMember struct {
	InstanceID int64
	Subject    string
	Body       []byte
}

// groupedReportAcc accumulates every instance of the same situation (same situation/rule/task
// triplet) that fired a "grouped" situation-reporting alert during a single scheduler run.
type groupedReportAcc struct {
	task    SituationReportingTask // as seen on the first member; GroupSubject/Timeout/IssueID are expected to be identical across members
	key     string
	to      []string
	cc      []string
	members []groupedReportMember
}

// buildGroupTaskKey builds the key identifying a group of alerts: same situation, same rule, same
// task id. Unlike buildSituationReportingKey, it deliberately drops the template instance id - that
// is the whole point of grouping: every instance of the situation shares this one key, so they all
// land in the same accumulator and get flushed as a single email.
func buildGroupTaskKey(input ContextData, task Task) string {
	return fmt.Sprintf("group-%d-%d-%s", input.SituationID, input.RuleID, task.GetID())
}

// collectGroupedSituationReporting renders one instance's report and stashes it into pending
// instead of sending it. Called once per (situation instance, run) for every SituationReportingTask
// configured with groupBySituation=true.
func collectGroupedSituationReporting(pending map[string]*groupedReportAcc, task SituationReportingTask, context ContextData) error {
	key := buildGroupTaskKey(context, task)
	memberKey := fmt.Sprintf("%s-%d", key, context.TemplateInstanceID)

	skip, err := task.isIssueAlreadyHandled(memberKey)
	if err != nil {
		return err
	}
	if skip {
		zap.L().Debug("Grouped SituationReportingTask member skipped - open/draft issue already existed",
			zap.String("key", key), zap.Int64("templateInstanceID", context.TemplateInstanceID))
		return nil
	}

	body, _, err := task.buildReportContent(context)
	if err != nil {
		return err
	}

	acc, ok := pending[key]
	if !ok {
		acc = &groupedReportAcc{task: task, key: key}
		pending[key] = acc
	}
	acc.to = unionStrings(acc.to, task.To)
	acc.cc = unionStrings(acc.cc, task.Cc)
	acc.members = append(acc.members, groupedReportMember{
		InstanceID: context.TemplateInstanceID,
		Subject:    task.Subject,
		Body:       body,
	})

	return nil
}

// flushGroupedSituationReporting sends one merged email per accumulated group. Called once, after
// every batch of the scheduler run has been processed by ApplyTasks.
//
// Note: attachments are not supported on grouped reports yet (each instance may reference a
// different fact export; merging several CSV attachments into one is left for a later iteration).
func flushGroupedSituationReporting(pending map[string]*groupedReportAcc) {
	for _, acc := range pending {
		if len(acc.members) == 0 {
			continue
		}

		timeoutDuration, err := time.ParseDuration(acc.task.Timeout)
		if err != nil {
			zap.L().Warn("Invalid timeout on grouped SituationReportingTask", zap.String("key", acc.key), zap.Error(err))
			continue
		}
		if !verifyCache(acc.key, timeoutDuration) {
			zap.L().Debug("Grouped SituationReportingTask skipped - timeout not reached", zap.String("key", acc.key))
			continue
		}

		message := email2.NewMessage(acc.task.GroupSubject, "text/html", buildGroupedReportBody(acc.members))
		message.To = acc.to
		message.CC = acc.cc

		if err := email2.S().Send(message); err != nil {
			zap.L().Warn("Error while sending grouped SituationReportingTask", zap.String("key", acc.key), zap.Error(err))
			continue
		}
		zap.L().Info("Grouped email sent !", zap.String("key", acc.key), zap.Int("membersCount", len(acc.members)))
	}
}

// buildGroupedReportBody concatenates every member's subject and rendered body into a single HTML
// body, clearly stating upfront that this message groups several alerts.
func buildGroupedReportBody(members []groupedReportMember) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<p><strong>Ce message regroupe %d alertes de m&ecirc;me situation.</strong></p>", len(members)))
	for _, m := range members {
		sb.WriteString("<hr/>")
		sb.WriteString(fmt.Sprintf("<h3>%s</h3>", html.EscapeString(m.Subject)))
		sb.Write(m.Body)
	}
	return sb.String()
}

// unionStrings appends to base every value of additional not already present in base, preserving
// order and without duplicates.
func unionStrings(base []string, additional []string) []string {
	seen := make(map[string]struct{}, len(base))
	result := make([]string, 0, len(base)+len(additional))
	for _, v := range base {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	for _, v := range additional {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}
