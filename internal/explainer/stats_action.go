package explainer

import "github.com/myrteametrics/myrtea-sdk/v5/postgres"

// ActionStats is a single action usage stats
type ActionStats struct {
	RootCauseID int64
	ActionID    int64
	Occurrences int64
	IssueCount  int64
}

// NewActionStats returns a new ActionStats
func NewActionStats(rootCauseID int64, actionID int64, occurrences int64, issueCount int64) ActionStats {
	return ActionStats{
		RootCauseID: rootCauseID,
		ActionID:    actionID,
		Occurrences: occurrences,
		IssueCount:  issueCount,
	}
}

// GetRelativeFrequency calculate the relative usage frequency of an action
func (as *ActionStats) GetRelativeFrequency() float64 {
	return float64(as.Occurrences) / float64(as.IssueCount)
}

// GetActionStats returns all actions usage statistics on a single situation
func GetActionStats(situationID int64) (map[int64]ActionStats, error) {
	query := `select
			ir.rootcause_id, 
			ir.action_id, 
			count(ir.action_id) as action_occurrences, 
			(select count(distinct(issue_id)) from issue_resolution_v1) as count_issues
		from issue_resolution_v1 ir 
		inner join issues_v1 i on ir.issue_id = i.id
		where i.situation_id = :situation_id
		group by ir.rootcause_id, ir.action_id`
	params := map[string]interface{}{
		"situation_id": situationID,
	}

	rows, err := postgres.DB().NamedQuery(query, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := make(map[int64]ActionStats, 0)
	for rows.Next() {
		var rootCauseID, actionID, actionOccurrences, countIssues int64
		err := rows.Scan(&rootCauseID, &actionID, &actionOccurrences, &countIssues)
		if err != nil {
			return nil, err
		}

		actionScore := NewActionStats(rootCauseID, actionID, actionOccurrences, countIssues)
		scores[actionID] = actionScore
	}
	return scores, nil
}
