package log

import (
	"bufio"
	"io"
)

// FlushRoutine continuously writes everything in given ReadCloser
// to an io.Writer. Use this as a goroutine.
func FlushRoutine(w io.Writer, rc io.ReadCloser, stop chan struct{}) {
	reader := bufio.NewReader(rc)
ROUTINE:
	for {
		select {
		case <-stop:
			line, err := reader.ReadBytes('\n')
			if err == nil {
				w.Write(line)
			}
			break ROUTINE
		default:
			// Read from pipe then write to ResponseWriter and flush it,
			// sending the copied content to the client.
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break ROUTINE
			}
			w.Write(line)
		}
	}
}
