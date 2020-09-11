package assistant

// Assistant is the interface that we're exposing as a plugin.
type Assistant interface {
	SentenceProcess(string, string, [][]string) ([]byte, []string, error)
}
