# Go UDP Doom Stream

Stream terminal output from `doom-ascii` over UDP.

This project runs as either a **server** or **client**:
- The server launches `doom-ascii` in a PTY, captures its live frame output, run-length encodes it, and sends it over UDP.
- The client connects, decodes frames, and renders the stream directly to stdout.

By default, both sides communicate on `127.0.0.1:4040`.

## Features

- Two startup modes: `server` and `client`
- Simple UDP handshake (`client-hello`)
- PTY-based capture of `doom-ascii` output
- Lightweight RLE compression/decompression for frame transport
- Graceful session shutdown via `exit` signal

## Project Layout

- `asciiServer.go` - main entrypoint, UDP server/client logic, and RLE codec
- `go.mod` - module metadata and dependencies
- `doom/doom-ascii` - bundled Doom ASCII renderer binary
- `doom/DOOM1.WAD` - Doom WAD data used by the server

## Requirements

- Go `1.25.6`+
- A terminal that can handle high-volume ANSI/text updates
- `doom/doom-ascii` and a valid WAD file (`DOOM1.wad` or `DOOM1.WAD`)
- 

## Quick Start

Open two terminals in this directory.

1) Start the server:

```bash
./asciiServer -mode=server
```
or
```bash
go run asciiServer.go -mode=server
```

2) Start the client:

```bash
./asciiServer -mode=client
```
or
```bash
go run asciiServer.go -mode=client
```

3) You can also launch without flags and choose interactively:

```bash
./asciiServer
```
or
```bash
go run asciiServer.go
```

## How It Works

1. Client sends a `client-hello` packet.
2. Server accepts the first client, starts `doom-ascii`, and reads PTY output.
3. Server RLE-encodes frame chunks and sends them over UDP.
4. Client decodes each payload and writes the frame data to stdout.
5. When the stream ends, server sends `exit` and the client returns.

## Notes

- Current bind/connect target is hardcoded to `127.0.0.1:4040` in `main()`.
- This is currently single-client per server run (first handshake wins).
- UDP is used as-is (no retransmit/ordering guarantees).
- This project was inspired by ThePrimeagen's video "1000 Players - One Game of Doom"
- This project relies on the github project doom-ascii by wojciech-graj
- Because of its reliance on the doom-ascii, a doom .WAD file is required and can be obtained via shareware
- The .WAD file should be named DOOM1.WAD and placed in the /doom/ directory
