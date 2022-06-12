package chaos

import (
	"encoding/json"
	"fmt"
)

type Method string

const (
	MethodIntroduction         Method = "introduction"
	MethodIntroductionResponse Method = "introductionResponse"
	MethodOkResponse           Method = "okResponse"
)

type Message struct {
	Method Method          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type MessageNodeIntroduction struct {
	Addr string `json:"addr"`
}

type MessageNodeIntroductionResponse struct {
	Addr  string   `json:"addr"`
	First bool     `json:"first"`
	Nodes []string `json:"nodes"` // TODO - Nodes is temporary until I decide what node structure I want to use
}

type MessageOkResponse struct {
	OK bool `json:"ok"`
}

func DecodeMessage(message Message) (any, error) {
	switch message.Method {
	case MethodIntroduction:
		var msg MessageNodeIntroduction
		return msg, json.Unmarshal(message.Params, &msg)
	default:
		return nil, fmt.Errorf("unknown method: %q", message.Method)
	}
}

func EncodeMessage(msg any) ([]byte, error) {
	message := struct {
		Method Method `json:"method"`
		Params any    `json:"params"`
	}{
		Method: getMessageMethod(msg),
		Params: msg,
	}
	return json.Marshal(message)
}

func getMessageMethod(msg any) Method {
	switch msg.(type) {
	case MessageNodeIntroduction:
		return MethodIntroduction
	case MessageNodeIntroductionResponse:
		return MethodIntroductionResponse
	case MessageOkResponse:
		return MethodOkResponse
	default:
		return ""
	}
}
