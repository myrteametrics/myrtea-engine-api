package coordinator

// RollMode is a enumeration of all indices roll mode
type RollMode int

const (
	// Rollover is the native elasticsearch roll mode, based on /_rollover API
	Rollover RollMode = iota + 1
	// IngestTimebased is ...
	IngestTimebased
)
