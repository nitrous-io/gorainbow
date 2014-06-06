package utils

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

func ShellEscape(src string) string {
	esc := shellRegexp.ReplaceAll(
		[]byte(strings.Replace(src, "\n", "\\n", -1)),
		[]byte("\\$1"))
	return string(esc)
}

func QuoteEscape(name string) string {
	esc := strings.Replace(name, `'`, `'"'"'`, -1)
	return fmt.Sprintf(`'%s'`, esc)
}

func EscapeArg(arg string) string {
	return QuoteEscape(ShellEscape(arg))
}

func MakeCallbackMsg(id int, err error, args ...interface{}) *protocol.Message {
	var cb []interface{}
	if err != nil {
		cb = []interface{}{err.Error()}
	} else {
		cb = []interface{}{nil}
	}
	cb = append(cb, args...)
	return protocol.CallbackMsg(id, cb)
}

// if the path is not absolute, start from the current directory
func GetPath(name string) string {
	if len(name) > 0 && name[0] == '/' {
		return name
	}
	return "./" + name
}

func WriteOne(ws sockjs.Writer, msg *protocol.Message) {
	ws.Write(protocol.EncodeServerMsg(msg))
}
