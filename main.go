package main

import (
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

func isExchangeConsidered(exchange string) bool {
	for _, x := range exchanges {
		if x == exchange {
			return true
		}
	}
	return false
}

func isPairConsidered(pair string) bool {
	for _, x := range pairs {
		if x == pair {
			return true
		}
	}
	return false
}

func recordMetrics() {
	go func() {
		for {
			restclient := rest.NewCWRESTClient(nil)

			marketSummaries, err := restclient.GetMarketSummaries()

			if err != nil {
				log.Fatal("Unable to fetch summaries")
			}

			for market, summary := range marketSummaries {
				r := strings.Split(market, ":")
				exchange, pair := r[0], r[1]
				if isExchangeConsidered(exchange) && isPairConsidered(pair) {
					last, _ := strconv.ParseFloat(summary.Last, 64)
					lastValue.WithLabelValues(exchange, pair).Set(last)
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
	exchanges = []string{"bitstamp", "kraken", "coinbase-pro"}
	pairs     = []string{"btcusd", "ltcusd"}
)

func main() {
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8899", nil)
}
