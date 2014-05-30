package protocol

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Namespace  string
	Event      string
	CallbackId int
	Args       []interface{}
}

func ErrorMsg(err error) *Message {
	return &Message{
		Namespace: "global",
		Event:     "error",
		Args:      []interface{}{fmt.Sprintf("%v", err)},
	}
}

func ParseClientMsg(data []byte) (msg *Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			msg = nil
			err = fmt.Errorf("Parse: panic %v", r)
		}
	}()
	generalMsg := []interface{}{}
	if err := json.Unmarshal(data, &generalMsg); err != nil {
		return nil, err
	}
	namespace := maybeString(generalMsg[0])
	event := maybeString(generalMsg[1])
	args := maybeArray(generalMsg[2])
	callback := 0
	if len(generalMsg) > 3 {
		callback = int(maybeFloat(generalMsg[3]))
	}
	return &Message{Namespace: namespace, Event: event, CallbackId: callback,
		Args: args}, nil
}

func EncodeServerMsg(msg *Message) []interface{} {
	e := []interface{}{}
	e = append(e, msg.Namespace)
	e = append(e, msg.Event)
	if msg.CallbackId == 0 {
		e = append(e, msg.Args)
	} else {
		e = append(e, []interface{}{msg.CallbackId, msg.Args})
	}
	return e
}

func maybeString(a interface{}) string {
	if a == nil {
		return ""
	}
	return a.(string)
}

func maybeFloat(a interface{}) float64 {
	if a == nil {
		return 0
	}
	return a.(float64)
}

func maybeArray(a interface{}) []interface{} {
	if a == nil {
		return []interface{}{}
	}
	return a.([]interface{})
}
