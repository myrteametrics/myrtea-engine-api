package model

type BoostInfo struct {
	JobID     string `json:"jobId"`
	Active    bool   `json:"active"`
	Frequency string `json:"frequency"`
	Quota     int    `json:"quota"`
	Used      int    `json:"used"`
}
