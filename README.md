# btenl ( Byte Tunnel )

## Overview

`btenl` is a command-line client, `btenld` is the background daemon. The client sends commands over a unix socket (`/tmp/btenld.sock`) and the daemon answers with an event-driven, worker-pool architecture — typed event channels, no shared-state contention.

## Architecture

```
  btenl (CLI)
    │  unix socket /tmp/btenld.sock
    ▼
  btenld (daemon)
    │
    ├── controls source ──► control events ──► worker pool
    ├── connections mux ──► daemon events ──► worker pool
    └── error events ─────► error worker ──► log
```

## Getting started

```bash
# build both binaries
go build ./cmd/...

# start the daemon
./btenl start

# send a command
./btenl <command>

# stop the daemon
./btenl stop
```

## Usage

| Command              | What it does                          |
| -------------------- | ------------------------------------- |
| `btenl start`        | start `btenld` in the background      |
| `btenl stop`         | stop the running daemon               |
| `btenl help`         | show help                             |
| `btenl version`      | print version                         |
| `btenl <command>`    | forward anything else to the daemon   |

## Layout

```
cmd/btenl      CLI client
cmd/btenld     daemon
internal/controls   control event sources (unix IPC)
internal/daemon     daemon core: sinks, workers, handlers
internal/logger     logging
internal/types      events, sinks, sources
```
