# dhtc-client

[dhtc](https://github.com/nbdy/dhtc) crawler part as a http client<br>
use [dhtc-server](https://github.com/nbdy/dhtc-server) with this

## usage

```shell
nbdy@c0:~$ ./dhtc-client --help
Usage of ./dhtc-client:
  -auth
        enable authentication
  -drainTimeout duration
        drain timeout (default 5s)
  -host string
        dhtc-server address (default "127.0.0.1")
  -maxLeeches int
        max. leeches (default 256)
  -maxNeighbors uint
        max. indexer neighbors (default 1000)
  -name string
        crawler name
  -port int
        dhtc-server port (default 7331)
  -protocol string
        protocol (default "http")
  -token string
        crawler token
```
