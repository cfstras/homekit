/*
https://manual.pilight.org/en/installation
*/
package pilight

import (
	"encoding/json"
	"log"
	"net"
)

// {\"action\":\"send\",
// \"code\":{
//    \"systemcode\":25,
//    \"unitcode\":4,
//    \"on\":1,
//    \"protocol\":[\"elro_800_switch\"]}}

type Message struct {
	Action Action `json:"action"`
	Code   Code   `json:"code"`
}
type Code struct {
	SystemCode int `json:"systemcode"`
	UnitCode   int `json:"unitcode"`

	// Set on=1 or off=1
	On  int8 `json:"on,omitempty"`
	Off int8 `json:"off,omitempty"`

	Protocol []string `json:"protocol"`
}

type Action string

var (
	ActionSend         Action = "send"
	ProtocolElroSwitch string = "elro_800_switch"
)

func (m Message) Send() {
	b, err := json.Marshal(&m)
	p(err, "toJson failed:", m)

	conn, err := net.Dial("tcp", "127.0.0.1:42777")
	if err != nil {
		log.Println("Switch failed:", conn)
		return
	}
	if _, err = conn.Write(b); err != nil {
		log.Println("Switch failed:", conn)
		return
	}
	if err = conn.Close(); err != nil {
		log.Println("Switch failed:", conn)
		return
	}
}
func p(err error, msg string, args ...interface{}) {
	if err != nil {
		args2 := append([]interface{}{"error", msg, err}, args...)
		log.Fatalln(args2)
	}
}
