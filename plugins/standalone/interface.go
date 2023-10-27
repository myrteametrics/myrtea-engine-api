package standalone

// StandaloneService
// A Standalone plugin will wait until the engine sends a Run
// The plugin timeouts after a given time and shutdown's afterward
type StandaloneService interface {
	Run(port int) error // Run's the plugin with parameters
}
