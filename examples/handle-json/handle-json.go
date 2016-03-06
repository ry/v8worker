package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ry/v8worker"
)

func DiscardSendSync(msg string) string { return "" }

type message struct {
	MessageType string      `json:"messageType"`
	Content     interface{} `json:"content,omitempty"`
	Handled     bool        `json:"handled"`
}

func runWorker(in <-chan message) <-chan message {
	out := make(chan message)
	go func() {
		worker := v8worker.New(func(msg string) {
			var message message
			if err := json.Unmarshal([]byte(msg), &message); err == nil {
				out <- message
			}
		}, DiscardSendSync)
		file, _ := ioutil.ReadFile("handle-json.js")
		worker.Load("handle-json.js", string(file))
		for m := range in {
			bytes, err := json.Marshal(m)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			msg := string(bytes)
			if err := worker.Send(msg); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
		}
		close(out)
		worker.TerminateExecution()
	}()
	return out
}

func main() {
	messages := []message{message{
		MessageType: "msg",
		Content:     "foo",
	}, message{
		MessageType: "msg",
		Content:     "bar",
	}}

	out := make(chan message)
	in := runWorker(out)
	go func() {
		for _, m := range messages {
			out <- m
		}
		close(out)
	}()
	for m := range in {
		fmt.Printf("got message: %v\n", m)
	}
}
