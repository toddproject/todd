package testlets

// Testlet defines what a testlet should look like if built in native
// go and compiled with the agent
type Testlet interface {
	Run(string, []string) ([]byte, error)
}
