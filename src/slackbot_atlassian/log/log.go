package log

import (
	"fmt"
	"os"
)

func LogF(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}
