# Dictago

Dictago is a Redis-like server implementation written in Go. It aims to provide a subset of Redis's functionality, focusing on core commands, data structures, and replication.

## Features

Dictago supports the following Redis commands:

- **String:** `SET`, `GET`, `INCR`
- **List:** `LPUSH`, `RPUSH`, `LRANGE`, `LLEN`, `LPOP`, `BLPOP`
- **Stream:** `XADD`, `XRANGE`, `XREAD`
- **Transactions:** `MULTI`, `EXEC`, `DISCARD`
- **Connection:** `PING`, `ECHO`
- **Security:** `ACL`, `AUTH`
- **Server:** `INFO`, `TYPE`, `KEYS`
- **Replication:** `REPLCONF`, `PSYNC`, `WAIT`

## Getting Started

### Prerequisites

- Go 1.18 or higher

### Building

To build the server, clone the repository and run the following command:

```sh
go build
```

### Running the server

To run the server, use the following command:

```sh
./dictago [--port <port>] [--replicaof <master_host>:<master_port>]
```

- `--port`: The port to listen on. Defaults to `6379`.
- `--replicaof`: The master to replicate from. If specified, the server will start as a replica.

## Usage

You can connect to the server using any Redis client, such as `redis-cli`.

```sh
redis-cli -p <port>
```

Once connected, you can use any of the supported commands.

## Testing

To run the tests, use the following command:

```sh
go test -v ./tests/
```
