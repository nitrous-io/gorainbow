package command

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
)

var (
	shellRegexp *regexp.Regexp
)

func init() {
	shellRegexp = regexp.MustCompile(`([^A-Za-z0-9_\-.,:\/@\n*\?])`)
}

func toString(d []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return string(d), nil
}

func shellEscape(src string) string {
	esc := shellRegexp.ReplaceAll(
		[]byte(strings.Replace(src, "\n", "\\n", -1)),
		[]byte("\\$1"))
	return string(esc)
}

func quoteEscape(src string) string {
	esc := strings.Replace(src, `'`, `'"'"'`, -1)
	return fmt.Sprintf(`'%s'`, esc)
}

func escapeArg(arg string) string {
	return quoteEscape(shellEscape(arg))
}

func writeOne(ws sockjs.Writer, msg *protocol.Message) {
	ws.Write(protocol.EncodeServerMsg(msg))
}

func makeCallbackMsg(id int, err error, args ...interface{}) *protocol.Message {
	var cb []interface{}
	if err != nil {
		cb = []interface{}{err.Error()}
	} else {
		cb = []interface{}{nil}
	}
	cb = append(cb, args...)
	return protocol.CallbackMsg(id, cb)
}
