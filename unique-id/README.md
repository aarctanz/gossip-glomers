# Challenge #2: Unique ID Generation

Globally-unique IDs via [Twitter Snowflake](https://blog.twitter.com/engineering/en_us/a/2010/announcing-snowflake).

Workload runs with `--availability total` and a partition nemesis: an isolated
node must still serve `generate`. That rules out coordination at request time,
so the ID space is partitioned up front.

## Bit layout

```
 63     62 .......... 22   21 .... 12   11 .... 0
[sign]  [ms since epoch]   [ node ID ]  [  seq  ]
   1           41              10           12

(lastMs << 22) | (nodeID << 12) | seq
```

| Field     | Bits | Range                | Purpose                                   |
| --------- | ---: | -------------------- | ----------------------------------------- |
| sign      |    1 | always 0             | keeps the ID a non-negative `int64`       |
| timestamp |   41 | ~69.7 years          | ms since epoch `1288834974657` (2010-11-04) |
| node ID   |   10 | 1024 nodes           | numeric part of the Maelstrom node ID     |
| sequence  |   12 | 4096 IDs / ms / node | disambiguates within one millisecond      |

Uniqueness across nodes: the node-ID field differs, so two nodes can never emit
the same integer. Within a node: a mutex plus a never-decreasing
(timestamp, sequence) pair. No messages exchanged.

Custom 2010 epoch spends none of the 41-bit range on years already past —
exhaustion ~2080. Timestamp in the high bits makes IDs sort chronologically.

Node ID must fit in `[0, 1023]`; `idGen.init` rejects larger rather than letting
it overflow into the timestamp field.






## Running

```bash
go build -o unique-id .

../maelstrom/maelstrom test -w unique-ids --bin ./unique-id \
  --time-limit 30 --rate 1000 --node-count 3 \
  --availability total --nemesis partition
```

```
:workload {:valid? true, :attempted-count 27796, :duplicated-count 0}
:availability {:valid? true, :ok-fraction 1.0}
Everything looks good! ヽ(‘ー`)ノ
```

## Trade-offs

- **1024-node ceiling**, fixed by the field width. More needs 128-bit IDs (ULID)
  or dynamic node-ID assignment — which reintroduces coordination.
- **Node IDs must never be concurrently reused.** Maelstrom guarantees this; a
  production deployment must too.
- **Clock drift is unbounded.** IDs stay unique and per-node ordered, but become
  a less accurate record of when they were issued.
- **Not unguessable.** Sequential, and leaks a timestamp and node identity. Fine
  as internal keys, wrong as security tokens.
