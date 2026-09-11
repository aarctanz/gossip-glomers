# Gossip Glomers

Solutions to [Fly.io's Gossip Glomers](https://fly.io/dist-sys/), built on
[Maelstrom](https://github.com/jepsen-io/maelstrom) — a Jepsen-based workbench
that runs your binary as a cluster of nodes, drives it with a workload, injects
faults, and checks the history for correctness violations.


## Challenges

| # | Challenge | Directory | Workload | Status |
| - | --------- | --------- | -------- | ------ |
| 1 | [Echo](https://fly.io/dist-sys/1/) | [`echo/`](echo/) | `echo` | ✅ passing |
| 2 | [Unique ID Generation](https://fly.io/dist-sys/2/) | [`unique-id/`](unique-id/) | `unique-ids` | ✅ passing |

**1. Echo** — reply to `echo` with `echo_ok` carrying the same payload. Sets up
the shape of everything after: a handler per message type.

**2. Unique ID Generation** — globally-unique IDs, still served while totally
partitioned. Coordination at request time is impossible, so the ID space is
partitioned up front using Twitter Snowflake: 41 timestamp bits, 10 node bits,
12 sequence bits. Write-up: [`unique-id/README.md`](unique-id/README.md).

## Setup

Needs **Go** 1.21+ and a **JDK** (Maelstrom runs on the JVM), plus `graphviz`
and `gnuplot` for its plots.


## Running

Each challenge is its own Go module. Run from the repo root.

```bash
# Echo
(cd echo && go build -o echo .)
./maelstrom/maelstrom test -w echo --bin ./echo/echo \
  --node-count 1 --time-limit 10

# Unique IDs
(cd unique-id && go build -o unique-id .)
./maelstrom/maelstrom test -w unique-ids --bin ./unique-id/unique-id \
  --time-limit 30 --rate 1000 --node-count 3 \
  --availability total --nemesis partition
```

A passing run ends with:

```
Everything looks good! ヽ(‘ー`)ノ
```
