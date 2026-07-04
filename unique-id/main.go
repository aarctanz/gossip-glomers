package main

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type idGen struct {
	mu     sync.Mutex
	lastMs int64
	seq    uint16
}

func (g *idGen) next(nodeID int64) int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()
	for now <= g.lastMs {
		now = time.Now().UnixMilli()
	}

	if now == g.lastMs {
		g.seq = (g.seq + 1)
	} else {
		g.seq = 0
	}
	g.lastMs = now

	epoch := now - 1288834974657
	return (epoch << 21) | (nodeID << 16) | int64(g.seq)
}

func main() {
	var c idGen
	n := maelstrom.NewNode()

	n.Handle("generate", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		nodeId := n.ID()
		nid, err := strconv.Atoi(nodeId[1:])
		if err != nil {
			return err
		}

		id := c.next(int64(nid))

		body["type"] = "generate_ok"
		body["id"] = id

		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
