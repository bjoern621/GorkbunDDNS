package logger

import (
	"fmt"
	"log"

	"bjoernblessin.de/gorkbunddns/src/util/assert"
)

// Errorf prints an error message prefixed with "[ERROR] " and stops execution.
// The line after Errorf is never executed.
func Errorf(format string, v ...any) {
	log.Fatalf(fmt.Sprintf("[ERROR] %s", format), v...)
	assert.Never()
}

// Warnf prints a message prefixed with "[WARN] ".
func Warnf(format string, v ...any) {
	log.Printf(fmt.Sprintf("[WARN] %s", format), v...)
}
