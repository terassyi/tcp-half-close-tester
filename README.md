# tcp-half-close-tester

tcp-half-close-tester is the test server and client to check TCP half-close state.

## Usage

### Server

```
$ bin/tcp-half-close-tester server -h
Run the tcp-half-close-tester server

Usage:
  tcp-half-close-tester server [flags]

Flags:
  -c, --chunk int          Chunk size to write (default 1024)
  -f, --file string        File path to send
  -h, --help               help for server
  -l, --listen string      Listen address:port (default "0.0.0.0:4000")
      --log-level string   Log level(debug, info, warn, error) (default "info")
```

### Client

```
$ bin/tcp-half-close-tester client -h
Run the tcp-half-close-tester client

Usage:
  tcp-half-close-tester client [flags]

Flags:
  -b, --buf-size int             Buffer size to read (default 1024)
  -h, --help                     help for client
      --log-level string         Log level(debug, info, warn, error) (default "info")
  -r, --read-timeout duration    Read timeout to close connection (default 10m0s)
  -s, --server string            Server address to connect (default "127.0.0.1:4000")
  -w, --write-timeout duration   Write timeout to close connection (default 10m0s)
```

## How to use

1. Build the program

```
$ make build
```

2. Generate test data file

Default file size is 100M.
And file name is `data`.
```
$ make gen-file
```

You can change the file size and file name like following.
```
$ make gen-file SIZE=200m FILE=another-data
```

3. Run server

In this example, the server sends data every 10bytes from the file named `./data` that has 100M.
```
$ bin/tcp-half-close-tester server -f ./data -c 10
{"time":"2024-12-04T11:18:00.976289277Z","level":"INFO","msg":"new server","server":{"listen":"0.0.0.0:4000","file":"./data","chunk":10}}
{"time":"2024-12-04T11:18:00.976408579Z","level":"INFO","msg":"ready to send data","server":{"listen":"0.0.0.0:4000","size":104857600}}
{"time":"2024-12-04T11:18:00.976567782Z","level":"INFO","msg":"start the server","server":{"listen":"0.0.0.0:4000"}}
```

4. Run client

The client reads data from server but it has timeout that is 10s to write data.

In this log, we can find the client could read data after closing the connection from client-side.
```
$ bin/tcp-half-close-tester client -s localhost:4000 -w 10s
{"time":"2024-12-04T11:32:32.401116185Z","level":"INFO","msg":"new client","client":{"server":"localhost:4000","read-timeout":600000000000,"write-timeout":10000000000}}
{"time":"2024-12-04T11:32:32.401606085Z","level":"INFO","msg":"start to read","client":{"server":"localhost:4000"}}
{"time":"2024-12-04T11:32:42.401274039Z","level":"INFO","msg":"write timeout is expired","client":{"server":"localhost:4000","timeout":10000000000}}
{"time":"2024-12-04T11:33:29.327145376Z","level":"INFO","msg":"got EOF","client":{"server":"localhost:4000","total":104857600}}
{"time":"2024-12-04T11:33:29.327198176Z","level":"INFO","msg":"finish handling","client":{"server":"localhost:4000"}}
```

This is the server's log after finishing sending data.
```
$ bin/tcp-half-close-tester server -f ./data -c 10
{"time":"2024-12-04T11:32:18.475384729Z","level":"INFO","msg":"new server","server":{"listen":"0.0.0.0:4000","file":"./data","chunk":10}}
{"time":"2024-12-04T11:32:18.47563213Z","level":"INFO","msg":"ready to send data","server":{"listen":"0.0.0.0:4000","size":104857600}}
{"time":"2024-12-04T11:32:18.47579383Z","level":"INFO","msg":"start the server","server":{"listen":"0.0.0.0:4000"}}
{"time":"2024-12-04T11:32:32.401720685Z","level":"INFO","msg":"start to stream with chunk","server":{"listen":"0.0.0.0:4000","chunk":10}}
{"time":"2024-12-04T11:33:29.327049176Z","level":"INFO","msg":"finish streaming with chunk","server":{"listen":"0.0.0.0:4000","size":104857600}}
```


This is the packet flow.
We can observe TCP half-closing in this log.
```
sudo tcpdump -i lo 'tcp[tcpflags] & (tcp-syn|tcp-fin) != 0'
tcpdump: verbose output suppressed, use -v[v]... for full protocol decode
listening on lo, link-type EN10MB (Ethernet), snapshot length 262144 bytes
11:32:32.401545 IP localhost.51620 > localhost.4000: Flags [S], seq 3737635538, win 65495, options [mss 65495,sackOK,TS val 380935647 ecr 0,nop,wscale 7], length 0
11:32:32.401556 IP localhost.4000 > localhost.51620: Flags [S.], seq 57464518, ack 3737635539, win 65483, options [mss 65495,sackOK,TS val 380935647 ecr 380935647,nop,wscale 7], length 0
11:32:42.401321 IP localhost.51620 > localhost.4000: Flags [F.], seq 1, ack 18762191, win 8698, options [nop,nop,TS val 380945647 ecr 380945647], length 0
11:33:29.327108 IP localhost.4000 > localhost.51620: Flags [F.], seq 104857601, ack 2, win 512, options [nop,nop,TS val 380992573 ecr 380992573], length 0
^C
4 packets captured
8 packets received by filter
0 packets dropped by kernel
```
