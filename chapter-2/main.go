package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	s := server{maelstrom.NewNode()}
	s.Run()
}

type server struct {
	n *maelstrom.Node
}

func (s server) Run() {
	s.n.Handle("generate", s.HandleGenerate)

	if err := s.n.Run(); err != nil {
		log.Fatal(err)
	}
}

func (s server) HandleGenerate(msg maelstrom.Message) error {
	var body map[string]any

	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	body["type"] = "generate_ok"
	body["id"] = fmt.Sprintf("%v%v", time.Now().UnixNano(), randInt(100000000, 999999999))

	return s.n.Reply(msg, body)
}

func randInt(low, hi int) int {
	return low + rand.Intn(hi-low)
}
