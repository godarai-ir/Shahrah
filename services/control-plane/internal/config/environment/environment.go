// Package environment owns direct access to process environment variables.
package environment

import "os"

// Environment returns one process environment value at the configuration edge.
func Environment(key string) string {
	return os.Getenv(key)
}

// LookupEnvironment returns one process environment value and whether it is set.
func LookupEnvironment(key string) (string, bool) {
	return os.LookupEnv(key)
}
