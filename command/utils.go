package command

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/nitrous-io/gorainbow/protocol"
	"github.com/nitrous-io/gorainbow/sockjs"
)

const (
	currentDirectory = "/home/action"
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
	return fmt.Sprintf(`'%s'`, escapeName(src))
}

func escapeName(name string) string {
	return strings.Replace(name, `'`, `'"'"'`, -1)
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

func getPath(name string) string {
	if len(name) > 0 {
		name = path.Join(currentDirectory, escapeName(name))
	} else {
		name = currentDirectory
	}
	return name
}

func regexTerm(term string) string {
	chars := strings.Split(term, "")
	return ".*?" + strings.Join(chars, ".*?") + ".*?"
}
