package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boramalper/magnetico/cmd/magneticod/bittorrent/metadata"
	"github.com/boramalper/magnetico/cmd/magneticod/dht"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Configuration struct {
	protocol string
	host     string
	port     int

	maxNeighbors uint
	maxLeeches   int
	drainTimeout time.Duration
}

type AddInfoHashResponse struct {
	error bool
}

var config = Configuration{}
var client = http.Client{}

func parseArguments() {
	flag.StringVar(&config.protocol, "protocol", "http", "protocol")
	flag.StringVar(&config.host, "host", "127.0.0.1", "dhtc-server address")
	flag.IntVar(&config.port, "port", 7331, "dhtc-server port")

	flag.UintVar(&config.maxNeighbors, "maxNeighbors", 1000, "max. indexer neighbors")
	flag.IntVar(&config.maxLeeches, "maxLeeches", 256, "max. leeches")
	flag.DurationVar(&config.drainTimeout, "drainTimeout", 5*time.Second, "drain timeout")

	flag.Parse()
}

func getBaseUrl() string {
	return config.protocol + "://" + config.host + ":" + strconv.Itoa(config.port) + "/"
}

func sendInfoHash(md metadata.Metadata, result AddInfoHashResponse) error {
	data, err := json.Marshal(md)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", getBaseUrl()+"api/v1/add", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	body, _ := ioutil.ReadAll(rsp.Body)
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
				fmt.Printf("Sent: %s\n", md.Name)
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
