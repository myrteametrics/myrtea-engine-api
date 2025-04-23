package models

// Ids of Issues to Close
type IssuesIdsToClose struct {
	Ids []int64 `json:"ids"`
}

// reason to close an issue
type Reason struct {
	S string `json:"reason"`
}

//  status
type CloseIssuesStatus struct {
	ErrorMessages string
	AllOk         bool
	SuccessCount  int
}
