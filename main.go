package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boramalper/magnetico/cmd/magneticod/bittorrent/metadata"
	"github.com/boramalper/magnetico/cmd/magneticod/dht"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Configuration struct {
	protocol string
	host     string
	port     int

	auth  bool
	name  string
	token string

	maxNeighbors uint
	maxLeeches   int
	drainTimeout time.Duration
}

type AddInfoHashResponse struct {
	error bool
	text  string
}

type AddTorrentRequest struct {
	Torrent metadata.Metadata
}

type AuthenticatedRequest struct {
	Name  string
	Token string

	Torrent metadata.Metadata
}

var config = Configuration{}
var client = http.Client{}

func parseArguments() {
	flag.StringVar(&config.protocol, "protocol", "http", "protocol")
	flag.StringVar(&config.host, "host", "127.0.0.1", "dhtc-server address")
	flag.IntVar(&config.port, "port", 7331, "dhtc-server port")

	flag.BoolVar(&config.auth, "auth", false, "enable authentication")
	flag.StringVar(&config.name, "name", "", "crawler name")
	flag.StringVar(&config.token, "token", "", "crawler token")

	flag.UintVar(&config.maxNeighbors, "maxNeighbors", 1000, "max. indexer neighbors")
	flag.IntVar(&config.maxLeeches, "maxLeeches", 256, "max. leeches")
	flag.DurationVar(&config.drainTimeout, "drainTimeout", 5*time.Second, "drain timeout")

	flag.Parse()
}

func getBaseUrl() string {
	return config.protocol + "://" + config.host + ":" + strconv.Itoa(config.port) + "/"
}

func makeRequestData(md metadata.Metadata) ([]byte, error) {
	if config.auth {
		return json.Marshal(AuthenticatedRequest{config.name, config.token, md})
	} else {
		return json.Marshal(AddTorrentRequest{md})
	}
}

func sendInfoHash(md metadata.Metadata, result AddInfoHashResponse) error {
	data, err := makeRequestData(md)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", getBaseUrl()+"api/torrent/add", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	body, _ := io.ReadAll(rsp.Body)
	return json.Unmarshal(body, &result)
}

func crawl() {
	indexerAddrs := []string{"0.0.0.0:0"}
	interruptChan := make(chan os.Signal, 1)

	trawlingManager := dht.NewManager(indexerAddrs, 1, config.maxNeighbors)
	metadataSink := metadata.NewSink(config.drainTimeout, config.maxLeeches)

	addResult := AddInfoHashResponse{}

	for stopped := false; !stopped; {
		select {
		case result := <-trawlingManager.Output():
			metadataSink.Sink(result)

		case md := <-metadataSink.Drain():
			err := sendInfoHash(md, addResult)
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
			} else {
				if addResult.error {
					fmt.Printf("Error: %s\n", addResult.text)
					trawlingManager.Terminate()
					stopped = true
				} else {
					fmt.Printf("Sent: %s\n", md.Name)
				}
			}
		case <-interruptChan:
			trawlingManager.Terminate()
			stopped = true
		}
	}
}

func main() {
	parseArguments()
	crawl()
}
