package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	s := newServer()
	s.Run()
}

type server struct {
	n          *maelstrom.Node
	values     []int
	valuesLock sync.RWMutex

	topo     map[string][]string
	topoLock sync.RWMutex

	failed chan failedMessage
}

func newServer() *server {
	return &server{
		n:      maelstrom.NewNode(),
		failed: make(chan failedMessage, 100),
	}
}

func (s *server) Run() {
	s.n.Handle("broadcast", s.HandleBroadcast)
	s.n.Handle("read", s.HandleRead)
	s.n.Handle("topology", s.HandleTopo)

	go func() {
		for {
			failed := <-s.failed
			var msg maelstrom.Message
			var err error

            fmt.Fprintf(os.Stdout, "Trying to republish %v", failed)

			if msg, err = s.n.SyncRPC(context.Background(), failed.target, failed.msg); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to publish to node %s with message %v", failed.target, failed.msg)
				s.failed <- failed
			}

			rspData, err := readFromMessage(msg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[Republish] failed to read response %v", err)
				s.failed <- failed
			}

			if reflect.ValueOf(rspData["type"]).String() != "broadcast_ok" {
				s.failed <- failed
			}

            fmt.Fprintf(os.Stderr, "Successfully republish")
		}
	}()

	if err := s.n.Run(); err != nil {
		log.Fatal(err)
	}
}

func (s *server) Reply(msg maelstrom.Message, data map[string]any) error {
	return s.n.Reply(msg, data)
}

func (s *server) HandleBroadcast(msg maelstrom.Message) error {
	data, err := readFromMessage(msg)
	if err != nil {
		return err
	}

	if message, ok := data["message"]; !ok {
		return errors.New("invalid message")
	} else {

		newValue := int(message.(float64))

		s.valuesLock.Lock()
		s.values = append(s.values, newValue)
		s.valuesLock.Unlock()

		adjacency := s.adjacencyNodes()
		go func(nodes []string, exclude string, msg any) {
			for _, n := range nodes {
				if n == exclude {
					continue
				}

				if err := s.n.Send(n, msg); err != nil {
					s.failed <- failedMessage{n, msg}
				}
			}
		}(adjacency, msg.Src, data)

		return s.Reply(msg, map[string]any{"type": "broadcast_ok"})
	}
}

type failedMessage struct {
	target string
	msg    any
}

func (s *server) HandleRead(msg maelstrom.Message) error {
	s.valuesLock.RLock()
	var currentValues []int
	for i := 0; i < len(s.values); i++ {
		currentValues = append(currentValues, s.values[i])
	}
	s.valuesLock.RUnlock()

	return s.Reply(msg, map[string]any{
		"type":     "read_ok",
		"messages": currentValues,
	})
}

type topology struct {
	Topology map[string][]string `json:"topology"`
}

func (s *server) HandleTopo(msg maelstrom.Message) error {
	var body topology
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.topoLock.Lock()
	s.topo = body.Topology
	s.topoLock.Unlock()

	return s.Reply(msg, map[string]any{"type": "topology_ok"})
}

func (s *server) adjacencyNodes() []string {
	s.topoLock.RLock()
	ns := s.topo[s.n.ID()]
	s.topoLock.RUnlock()

	return ns
}

func readFromMessage(msg maelstrom.Message) (map[string]any, error) {
	var body map[string]any

	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return nil, err
	}

	return body, nil
}
