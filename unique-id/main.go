package main

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const (
	timestampBits = 41
	nodeIDBits    = 10
	seqBits       = 12

	maxNodeID = (1 << nodeIDBits) - 1
	maxSeq    = (1 << seqBits) - 1

	nodeIDShift    = seqBits
	timestampShift = seqBits + nodeIDBits

	epochMs = 1288834974657
)

type idGen struct {
	mu     sync.Mutex
	nodeID int64

	// Wall clock can jump backwards (NTP steps, `date -s`, VM resume), which
	// would let us re-issue timestamps we have already used. So the wall clock
	// is sampled exactly once, at init, and every subsequent reading is derived
	// from the monotonic clock instead.
	startTime   time.Time
	startWallMs int64

	lastMs int64
	seq    int64
}

// init records the node's identity and pins the clock base. It must be called
// before next, and is called once from the "init" handler.
func (g *idGen) init(nodeID int64) error {
	if nodeID < 0 || nodeID > maxNodeID {
		return fmt.Errorf("node ID %d out of range [0,%d]", nodeID, maxNodeID)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.nodeID = nodeID
	g.startTime = time.Now()
	g.startWallMs = g.startTime.UnixMilli()
	g.lastMs = g.nowMs()
	g.seq = -1

	return nil
}

// nowMs returns the current time in milliseconds since the snowflake epoch
func (g *idGen) nowMs() int64 {
	return g.startWallMs - epochMs + time.Since(g.startTime).Milliseconds()
}

func (g *idGen) next() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := g.nowMs()
	if now > g.lastMs {
		g.lastMs = now
		g.seq = 0
	} else {
		g.seq++
		if g.seq > maxSeq {
			// if sequence exhausted, wait for next milli second
			for now <= g.lastMs {
				runtime.Gosched()
				now = g.nowMs()
			}
			g.lastMs = now
			g.seq = 0
		}
	}

	return (g.lastMs << timestampShift) | (g.nodeID << nodeIDShift) | g.seq
}

func parseNodeID(s string) (int64, error) {
	if len(s) < 2 || s[0] != 'n' {
		return 0, fmt.Errorf("unexpected node ID format: %q", s)
	}
	return strconv.ParseInt(s[1:], 10, 64)
}

func main() {
	var g idGen
	n := maelstrom.NewNode()

	n.Handle("init", func(msg maelstrom.Message) error {
		nodeID, err := parseNodeID(n.ID())
		if err != nil {
			return err
		}
		return g.init(nodeID)
	})

	n.Handle("generate", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type": "generate_ok",
			"id":   strconv.FormatInt(g.next(), 10),
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
