package containerd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
)

type data struct {
	Type string
	Data []byte
}

// decode output of stdstream from containerd into a collection of writers
func decodeStdstream(s io.Reader, p map[string]io.Writer) error {
	decoder := json.NewDecoder(s)
	for {
		d := &data{}
		if err := decoder.Decode(d); err != nil {
			log.Println("decode stream", err)
			return err
		}
		w, ok := p[d.Type]
		if !ok {
			continue
		}
		_, err := io.Copy(w, bufio.NewReader(bytes.NewBuffer(d.Data)))
		if err != nil {
			return err
		}
	}
	return nil
}
