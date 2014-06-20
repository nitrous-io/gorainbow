package containerd

import (
	"bytes"
	"encoding/json"
	"io"
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
			if err == io.EOF {
				break
			}
		}
		w, ok := p[d.Type]
		if !ok {
			continue
		}
		if _, err := io.Copy(w, bytes.NewBuffer(d.Data)); err != nil {
			return err
		}
	}
	return nil
}
