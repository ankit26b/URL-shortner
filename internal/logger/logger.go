package logger

import (
	"io"
	"log"
	"os"
)

func New(env string) *log.Logger {
	var output io.Writer = os.Stdout
	prefix := "[INFO] "
	if env == "production" {
		prefix = ""
	}
	return log.New(output, prefix, log.LstdFlags|log.LUTC|log.Lshortfile)
}
