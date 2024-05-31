package explainer

import (
	"github.com/myrteametrics/myrtea-sdk/v5/postgres"
)

// RootCauseStats is a single rootcause usage stats
type RootCauseStats struct {
	RootCauseID int64
	Occurrences int64
	IssueCount  int64
}

// NewRootCauseStats returns a new RootCauseStats
func NewRootCauseStats(rootCauseID int64, occurrences int64, issueCount int64) RootCauseStats {
	return RootCauseStats{
		RootCauseID: rootCauseID,
		Occurrences: occurrences,
		IssueCount:  issueCount,
	}
}

// GetRelativeFrequency calculate the relative usage frequency of a rootcause
func (as *RootCauseStats) GetRelativeFrequency() float64 {
	return float64(as.Occurrences) / float64(as.IssueCount)
}

// GetRootCauseStats returns all rootcauses usage statistics on a single situation
func GetRootCauseStats(situationID int64) (map[int64]RootCauseStats, error) {
	query := `select
			ir.rootcause_id,
			count(distinct(ir.issue_id, ir.rootcause_id)) as occurrences,
			(select count(distinct(issue_id)) from issue_resolution_v1) as count_issues
		from issue_resolution_v1 ir
		inner join issues_v1 i on ir.issue_id = i.id
		where i.situation_id = :situation_id
		group by ir.rootcause_id`
	params := map[string]interface{}{
		"situation_id": situationID,
	}

	rows, err := postgres.DB().NamedQuery(query, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := make(map[int64]RootCauseStats, 0)
	for rows.Next() {
		var rootCauseID, actionOccurrences, countIssues int64
		err := rows.Scan(&rootCauseID, &actionOccurrences, &countIssues)
		if err != nil {
			return nil, err
		}

		rootCauseScore := NewRootCauseStats(rootCauseID, actionOccurrences, countIssues)
		scores[rootCauseID] = rootCauseScore
	}
	return scores, nil
}
