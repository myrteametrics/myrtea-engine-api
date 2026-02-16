package model

type BoostInfo struct {
	JobID     string `json:"jobId"`
	Active    bool   `json:"active"`    // Is boost currently active?
	Frequency string `json:"frequency"` // Boost cron expression (e.g. "*/1 * * * *")
	Quota     int    `json:"quota"`     // Max executions in boost mode
	Used      int    `json:"used"`      // Executions completed in boost mode
}
