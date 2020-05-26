package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"code.cryptowat.ch/cw-sdk-go/client/rest"
)

func recordMetrics() {
	go func() {
		for {
			restclient := rest.NewCWRESTClient(nil)

			marketSummaries, err := restclient.GetMarketSummaries()

			if err != nil {
				log.Fatal("Unable to fetch summaries")
			}

			exchangesSlice := strings.Split(exchanges, ",")
			pairsSlice := strings.Split(pairs, ",")

			for _, exchange := range exchangesSlice {
				for _, pair := range pairsSlice {
					if summary, present := marketSummaries[fmt.Sprintf("%s:%s", exchange, pair)]; present {
						last, _ := strconv.ParseFloat(summary.Last, 64)
						lastValue.WithLabelValues(exchange, pair).Set(last)
					}
				}

			}

			time.Sleep(60 * time.Second)
		}
	}()
}

var (
	lastValue = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "crypto_last_value",
		Help: "The last known value in a given market (exchange/pair)",
	},
		[]string{
			"exchange",
			"pair",
		})
	exchanges     string
	pairs         string
	listenAddress string
)

func init() {
	flag.StringVar(&exchanges, "cryptowat.exchanges", "bitstamp,kraken,coinbase-pro", "Comma separated list of exchanges")
	flag.StringVar(&pairs, "cryptowat.pairs", "btcusd,ltcusd", "Comma separated list of pairs")
	flag.StringVar(&listenAddress, "web.listen-address", ":9150", "Address to listen on for web interface and telemetry")
	flag.Parse()
}

func main() {
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listenAddress, nil)
}
