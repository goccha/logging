package log

import (
	"io"
	"os"
)

func getWriter() (w io.Writer) {
	w = os.Stdout
	if v, ok := os.LookupEnv("LOGGING_OUT_FILE"); ok {
		if v != "" {
			if f, err := os.Create(v); err != nil {
				panic(err)
			} else {
				w = f
			}
		}
	}
	return
}

func getErrorWriter() (w io.Writer) {
	w = os.Stderr
	if v, ok := os.LookupEnv("LOGGING_ERROR_FILE"); ok {
		if v != "" {
			if f, err := os.Create(v); err != nil {
				panic(err)
			} else {
				w = f
			}
		}
	}
	return
}
