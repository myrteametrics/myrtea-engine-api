package model

// FrontDraft is a type alias used for FrontRecommendation persistence
// It allow an easier recognition inside the different functions to distinguish User Draft and generated Recommendation
type FrontDraft = FrontRecommendation

// FrontRecommendation is the main exchange structure with the frontend
// It is used to display recommendations, and to process issue resolving draft and feedback
type FrontRecommendation struct {
	ConcurrencyUUID string            `json:"uuid"`
	Tree            []*FrontRootCause `json:"tree"`
}

// FrontRootCause represent a single rootcause and its actions
type FrontRootCause struct {
	ID              int64          `json:"id"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	Selected        bool           `json:"selected"`
	Custom          bool           `json:"custom"`
	Occurrence      int64          `json:"occurrence"`
	UsageRate       float64        `json:"usageRate"`
	ClusteringScore float64        `json:"clusteringScore"`
	Actions         []*FrontAction `json:"actions"`
}

// FrontAction represent a single action
type FrontAction struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Selected    bool    `json:"selected"`
	Custom      bool    `json:"custom"`
	Occurrence  int64   `json:"occurrence"`
	UsageRate   float64 `json:"usageRate"`
}

// Ids of Issues to draft
type IssuesIdsToDraft struct {
	Ids     []int64 `json:"ids"`
	Comment *string `json:"comment,omitempty"`
}

// status
type DraftIssuesStatus struct {
	ErrorMessages string
	AllOk         bool
	SuccessCount  int
}
