package env

import (
	"os"

	"bjoernblessin.de/gorkbunddns/util/assert"
	"bjoernblessin.de/gorkbunddns/util/logger"
)

// ReadRequiredEnv reads the environment variable named by key and returns the read variable.
// Prints an error message and stops execution if it's not present.
// Returned string may be the empty string ("").
func ReadRequiredEnv(key string) string {
	env, present := os.LookupEnv(key)

	if !present {
		logger.Errorf("Environment variable %s not set. This is a required variable.", key)
		assert.Never()
	}

	return env
}

// ReadNonEmptyRequiredEnv acts like ReadRequiredEnv but fails if the variable is empty.
func ReadNonEmptyRequiredEnv(key string) string {
	env := ReadRequiredEnv(key)

	if env == "" {
		logger.Errorf("Environment variable %s is the empty string (\"\"). The variable must be non-empty.", key)
		assert.Never()
	}

	return env
}

// ReadOptionalEnv acts exactly like [os.LookupEnv].
func ReadOptionalEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}
