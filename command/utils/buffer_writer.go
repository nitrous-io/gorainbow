package utils

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"strings"
)

var ErrLimitLines = errors.New("LIMIT")

type BufferWriter struct {
	lineLimit int           // limit the number of lines that we scan
	outBuffer *bytes.Buffer // current output buffer waiting to be drained
	output    io.Writer     // current writer
	lineCount int           // current lines that has been read
}

func (c *BufferWriter) Write(p []byte) (n int, err error) {
	tobeWritten := bytes.NewBuffer(nil)
	tobeRead := bytes.NewBuffer(nil)
	// start copying our buffer and new data
	io.Copy(tobeRead, c.outBuffer)
	io.Copy(tobeRead, bytes.NewBuffer(p))
	// reset the buffer to prepare for new cached data
	c.outBuffer.Reset()

	var excess string

	// if the data doesnt contain newline, we want it to be written immediately
	// otherwise, just write til we cant find any newline and cache the
	// rest of the output
	if !strings.Contains(tobeRead.String(), "\n") {
		io.Copy(c.output, tobeRead)
	} else {
		scanner := bufio.NewScanner(tobeRead)
		scanner.Split(scanNewLines)
		for scanner.Scan() {
			t := scanner.Text()
			// if the previous line reached the limit, mark remainder as excess
			// and do not write
			if c.lineCount == c.lineLimit {
				excess = t
				break
			}
			if t[len(t)-1] == '\n' {
				tobeWritten.WriteString(t)
				c.lineCount++
			} else {
				c.outBuffer.Write([]byte(t))
			}
		}
		// if there are any error, just return it here. THis should not happen
		if err := scanner.Err(); err != nil {
			log.Printf("scannerOutput err: %v with data %v", err, string(p))
			return len(p), err
		}
		io.Copy(c.output, tobeWritten)
	}
	if c.lineLimit == 0 {
		return len(p), nil
	}
	if c.lineCount == c.lineLimit && excess != "" {
		return len(p), ErrLimitLines
	} else {
		return len(p), nil
	}
}

func NewBufferWriter(lineLimit int, output io.Writer) *BufferWriter {
	return &BufferWriter{
		lineLimit: lineLimit,
		outBuffer: bytes.NewBuffer(nil),
		output:    output,
	}
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

// scanNewLines is a split function for a Scanner that returns each line of
// text, include any newline marker `\n`. The returned line may
// be empty. The end-of-line marker is one optional carriage return followed
// by one mandatory newline. In regular expression notation, it is `\r?\n`.
// The last non-empty line of input will be returned even if it has no
// newline.
func scanNewLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0 : i+1]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}
