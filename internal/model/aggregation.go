package model

type JobBoostInfo struct {
	JobID      string `json:"jobId"`
	Configured bool   `json:"configured"` // Is boost configured for this job?
	Active     bool   `json:"active"`     // Is boost currently active?
	Frequency  string `json:"frequency"`  // JobBoostInfo cron expression (e.g. "*/1 * * * *")
	Quota      int    `json:"quota"`      // Max executions in boost mode
	Used       int    `json:"used"`       // Executions completed in boost mode
}
