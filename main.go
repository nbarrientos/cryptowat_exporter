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
		restclient := rest.NewCWRESTClient(nil)
		for {
			marketSummaries, err := restclient.GetMarketSummaries()

			if err != nil {
				log.Println("Unable to fetch summaries")
			}

			exchangesSlice := strings.Split(exchanges, ",")
			pairsSlice := strings.Split(pairs, ",")

			for _, exchange := range exchangesSlice {
				for _, pair := range pairsSlice {
					log.Printf("Looking for market %s:%s", exchange, pair)
					if summary, present := marketSummaries[fmt.Sprintf("%s:%s", exchange, pair)]; present {
						last, _ := strconv.ParseFloat(summary.Last, 64)
						lastValue.WithLabelValues(exchange, pair).Set(last)
						high, _ := strconv.ParseFloat(summary.High, 64)
						highValue.WithLabelValues(exchange, pair).Set(high)
						low, _ := strconv.ParseFloat(summary.Low, 64)
						lowValue.WithLabelValues(exchange, pair).Set(low)
						changeAbsolute, _ := strconv.ParseFloat(summary.ChangeAbsolute, 64)
						changeAbsoluteValue.WithLabelValues(exchange, pair).Set(changeAbsolute)
						changePercent, _ := strconv.ParseFloat(summary.ChangePercent, 64)
						changePercentValue.WithLabelValues(exchange, pair).Set(changePercent)
					}
				}

			}

			time.Sleep(time.Duration(cacheSeconds) * time.Second)
		}
	}()
}

var (
	lastValue = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crypto_last_value",
			Help: "The last known value in a given market (exchange/pair)",
		},
		[]string{
			"exchange",
			"pair",
		},
	)
	highValue = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crypto_high_24h_value",
			Help: "The 24h highest value in a given market (exchange/pair)",
		},
		[]string{
			"exchange",
			"pair",
		},
	)
	lowValue = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crypto_low_24h_value",
			Help: "The 24h lowest value in a given market (exchange/pair)",
		},
		[]string{
			"exchange",
			"pair",
		},
	)
	changePercentValue = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crypto_change_percent_value",
			Help: "The 24h percentage change in a given market (exchange/pair)",
		},
		[]string{
			"exchange",
			"pair",
		},
	)
	changeAbsoluteValue = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crypto_change_absolute_value",
			Help: "The 24h absolute change in a given market (exchange/pair)",
		},
		[]string{
			"exchange",
			"pair",
		},
	)
	exchanges     string
	pairs         string
	listenAddress string
	cacheSeconds  int
)

func init() {
	flag.StringVar(&exchanges, "cryptowat.exchanges", "bitstamp,kraken,coinbase-pro", "Comma separated list of exchanges")
	flag.StringVar(&pairs, "cryptowat.pairs", "btcusd,ltcusd", "Comma separated list of pairs")
	flag.IntVar(&cacheSeconds, "cryptowat.cacheseconds", 60, "Number of seconds to cache values for")
	flag.StringVar(&listenAddress, "web.listen-address", ":9150", "Address to listen on for web interface and telemetry")
	flag.Parse()
}

func main() {
	recordMetrics()
	log.Printf("Listening on address %s", listenAddress)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listenAddress, nil)
}
