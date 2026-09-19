# KV Store

> **Work in progress:** This is a learning project, built to explore
> distributed-systems concepts incrementally. The current implementation is a
> single-process HTTP key–value store, not a distributed database yet.

A small HTTP key–value store written in Go. It keeps active data in memory and
uses a write-ahead log (WAL) plus snapshots to restore state after a restart.

> **Revision note:** Sections marked **Your notes** are deliberately incomplete.
> Replace the prompts with your own explanations, design decisions, and lessons
> from building this project.

## Why I built this

- I built this project to explore distributed systems concepts which include
  - WAL - write-ahead logging (done)
  - State snapshot (done)
  - Implement leader-follower replication -- this is where "distributed" starts
  - inter-node communication using HTTP or RPC

## Project status

This repository is intentionally evolving. The implemented pieces establish a
single-node foundation API, concurrent access, durable mutation logging, and
recovery that can be used to investigate broader distributed-systems ideas.

| Area | Current state | Questions to explore |
| --- | --- | --- |
| Storage | In-memory string key–value map. | How should data be partitioned or replicated? |
| Durability | WAL records are synced; snapshots are written periodically. | What consistency guarantees does this provide during failures? |
| Recovery | Restores a snapshot, then replays later WAL records. | How should corrupted or partial records be recovered? |
| Concurrency | One process uses an `RWMutex` for its map. | What replaces a local lock across nodes? |
| Distribution | Not implemented. | How will nodes communicate, agree on order, and handle failures? |


## Features

- Create, read, update, delete, list, and clear string key–value pairs.
- JSON HTTP API served on port `8080`.
- Concurrent in-memory access protected with an `RWMutex`.
- Write-ahead logging for mutations once startup recovery is complete.
- Snapshotting to `snapshot.json` every 10 seconds.
- Startup recovery from the latest snapshot followed by WAL replay.

## Architecture

```text
HTTP client
    |
    v
handlers  --->  in-memory repository  --->  WAL (wal.log)
                         |
                         v
                  snapshotter (every 10s)
                         |
                         v
                 snapshot.json
```

### Components

| Location | Responsibility |  |
| --- | --- | --- |
| `cmd/api/main.go` | Wires dependencies, restores state, registers routes, and starts snapshotting. | <!-- Explain the startup order and why it matters. --> |
| `internal/handler` | Parses HTTP requests and returns JSON responses. | <!-- Explain the boundary between HTTP and storage logic. --> |
| `internal/repository` | Defines the store interface and implements the in-memory map. | <!-- Explain `RWMutex` and why reads/writes use different locks. --> |
| `internal/wal` | Serializes mutation records, appends and syncs them, then replays them. | <!-- Explain write-ahead logging in your own words. --> |
| `internal/snapshot` | Persists a point-in-time map plus its WAL checkpoint. | <!-- Explain what a checkpoint makes possible. --> |

## Data durability and recovery

At startup, the application:

1. Opens or creates `wal.log`.
2. Opens or creates `snapshot.json`.
3. Loads the latest snapshot state.
4. Replays WAL records from the snapshot checkpoint onward.
5. Enables WAL writes for new mutations.

For a normal mutation, the store writes and syncs a WAL record before changing
the in-memory map. Snapshots capture both the current map and the corresponding
WAL operation checkpoint.

<!--
Your notes:
- Why must the log be written before the in-memory change?
- Why is replay required even when a snapshot exists?
- Explain what happens after a crash at each point in a write.
- What assumptions does this design make? What failure cases are not handled?
-->
- The wal log must be written before the in-memory change happens so that the data can be recovered if the node crashes while making the change


## Getting started

### Prerequisites

- Go `1.25.0` or a compatible Go toolchain (see `go.mod`).
- `curl` is optional for trying the API examples.

### Run the server

```bash
go run ./cmd/api
```

The API is available at `http://localhost:8080`. Running it creates `wal.log`
and `snapshot.json` in the directory where the command is run.

### Run tests

```bash
go test ./...
```

## API

All responses use this shape:

```json
{
  "status": true,
  "msg": "...",
  "data": null
}
```

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/` | Health check. |
| `GET` | `/kv` | List all keys. |
| `GET` | `/kv/{key}` | Read a value. |
| `POST` | `/kv` | Create a key–value pair. |
| `PUT` | `/kv/{key}` | Update an existing value. |
| `DELETE` | `/kv/{key}` | Delete a key. |
| `DELETE` | `/kv` | Clear all keys. |

### Examples

```bash
# Health check
curl http://localhost:8080/

# Create a value
curl -X POST http://localhost:8080/kv \
  -H 'Content-Type: application/json' \
  -d '{"key":"language","value":"go"}'

# Read it
curl http://localhost:8080/kv/language

# Update it
curl -X PUT http://localhost:8080/kv/language \
  -H 'Content-Type: application/json' \
  -d '{"value":"golang"}'

# List keys, delete one, or clear the store
curl http://localhost:8080/kv
curl -X DELETE http://localhost:8080/kv/language
curl -X DELETE http://localhost:8080/kv
```

## WAL record format

Each log entry is a pipe-delimited line:

```text
<operation-id>|<operation>|<key>|<value>
```

Example:

```text
0|SET|language|go
```

Supported operations are `SET`, `UPDATE`, `DELETE`, and `CLEAR`.

<!--
Your notes:
- Why did you choose this format?
- What happens if a key or value contains `|` or a newline?
- How would you make parsing and recovery more robust?
-->
- `|` was choosen as a delimeter instead of empty space for easy parsing as `key` and `value` are allowed to contain empty spaces

## Snapshot format

Snapshots are stored as JSON with the map state and WAL checkpoint:

```json
{
  "wal_checkpoint": 0,
  "data": {
    "language": "go"
  }
}
```

<!--
Your notes:
- Why does a snapshot need both fields?
- How did you decide on the snapshot interval?
- What trade-off exists between snapshot frequency, recovery time, and I/O?
-->

## Things I learned

<!--
Write this section after revising the codebase.

- Go interfaces and dependency boundaries:
- `sync.RWMutex` and safe shared state:
- HTTP routing and JSON validation:
- WAL durability and `fsync`/`File.Sync`:
- Snapshotting and crash recovery:
- Tests I found most useful:
-->

## Current limitations and next steps

<!--
Your notes:
- This is a single-process, in-memory service; explain its persistence model.
- How should malformed or partially written WAL records be handled?
- Should log compaction/truncation be introduced after snapshots?
- What are the shutdown and snapshot-consistency considerations?
- Which distributed-systems concept will you explore next, and why?
- Possible directions: replication, leader election, consensus, sharding,
  membership, failure detection, client retries, or observability.
- For each new feature, define the guarantee it should provide and the failure
  cases it should tolerate.
-->

## Project structure

```text
.
├── cmd/api/main.go             # application entry point and lifecycle
├── internal/handler/           # HTTP handlers and response helpers
├── internal/repository/        # store interface and in-memory implementation
├── internal/wal/               # WAL record format, append, and replay
├── internal/snapshot/          # snapshot persistence and loading
└── README.md
```
