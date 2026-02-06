# Go ASCII Server

A lightweight UDP client/server program written in Go for sending plain-text (ASCII) messages over a local network.

## What it does

- Starts in `server` or `client` mode
- Uses UDP on `127.0.0.1:4040` by default
- Server accepts a client handshake, then sends messages from terminal input
- Client listens and prints incoming messages in real time
- Sending `exit` closes the session cleanly

## Project layout

- `rdt_packet.go` - main program entry and client/server logic
- `go.mod` - Go module definition

## Requirements

- Go `1.25.6` or newer

## Run the program

Open two terminals in the project folder.

### 1) Start the server

```bash
go run rdt_packet.go -mode=server
```

### 2) Start the client

```bash
go run rdt_packet.go -mode=client
```

You can also run without flags and choose interactively:

```bash
go run rdt_packet.go
```

## Example flow

1. Start server
2. Start client
3. Type messages in the server terminal
4. Read messages in the client terminal
5. Type `exit` on the server to close both sides

## Notes

- Current default address is local-only (`127.0.0.1`)
- Message buffer limit is `256` bytes per read
- This is a simple educational baseline for UDP communication
