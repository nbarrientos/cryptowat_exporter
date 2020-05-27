// Copyright 2020 Nacho Barrientos
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
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

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func recordMetrics(exchanges string, pairs string, cacheSeconds string) {
	go func() {
		restclient := rest.NewCWRESTClient(nil)
		for {
			marketSummaries, err := restclient.GetMarketSummaries()

			if err != nil {
				log.Println("Unable to fetch summaries")
			}

			exchangesSlice := strings.Split(exchanges, ",")
			pairsSlice := strings.Split(pairs, ",")
			var lastScrapeEpochMillis float64 = float64(time.Now().UnixNano()) / 1000

			for _, exchange := range exchangesSlice {
				for _, pair := range pairsSlice {
					log.Printf("Looking for market %s:%s", exchange, pair)
					if summary, present := marketSummaries[fmt.Sprintf("%s:%s", exchange, pair)]; present {
						cwLast, _ := strconv.ParseFloat(summary.Last, 64)
						last.WithLabelValues(exchange, pair).Set(cwLast)
						cwHigh24, _ := strconv.ParseFloat(summary.High, 64)
						high24.WithLabelValues(exchange, pair).Set(cwHigh24)
						cwLow24, _ := strconv.ParseFloat(summary.Low, 64)
						low24.WithLabelValues(exchange, pair).Set(cwLow24)
						cwChangeAbsolute, _ := strconv.ParseFloat(summary.ChangeAbsolute, 64)
						changeAbsolute.WithLabelValues(exchange, pair).Set(cwChangeAbsolute)
						cwChangePercent, _ := strconv.ParseFloat(summary.ChangePercent, 64)
						changePercent.WithLabelValues(exchange, pair).Set(cwChangePercent)
						lastUpdate.WithLabelValues(exchange, pair).Set(lastScrapeEpochMillis)
					} else {
						log.Printf("Unable to get information for market %s:%s", exchange, pair)
					}
				}
			}

			sleepSeconds, _ := strconv.ParseInt(cacheSeconds, 10, 32)
			log.Printf("Sleeping for %d seconds", sleepSeconds)
			time.Sleep(time.Duration(sleepSeconds) * time.Second)
		}
	}()
}

var (
	last = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crypto_currency",
			Help: "The last known trading value in a given market in the currency of the RHS of the pair",
		},
		[]string{
			"exchange",
			"pair",
		},
	)
	high24 = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crypto_high_24h_currency",
			Help: "The 24h highest value in a given market in the currency of the RHS of the pair",
		},
		[]string{
			"exchange",
			"pair",
		},
	)
	low24 = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crypto_low_24h_currency",
			Help: "The 24h lowest value in a given market in the currency of the RHS of the pair",
		},
		[]string{
			"exchange",
			"pair",
		},
	)
	changePercent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crypto_change_24h_ratio",
			Help: "The 24h change ratio in a given market",
		},
		[]string{
			"exchange",
			"pair",
		},
	)
	changeAbsolute = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crypto_change_24h_currency",
			Help: "The 24h absolute change in a given market in the currency of the RHS of the pair",
		},
		[]string{
			"exchange",
			"pair",
		},
	)
	lastUpdate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crypto_last_update_seconds",
			Help: "Seconds since epoch of last update",
		},
		[]string{
			"exchange",
			"pair",
		},
	)
)

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9745").String()
		exchanges     = kingpin.Flag("cryptowat.exchanges", "Comma separated list of exchanges.").Default("kraken,bitstamp").String()
		pairs         = kingpin.Flag("cryptowat.pairs", "Comma separated list of pairs.").Default("btcusd,ltcusd").String()
		cacheSeconds  = kingpin.Flag("cryptowat.cacheseconds", "Number of seconds to cache values for.").Default("60").String()
	)
	kingpin.Parse()

	recordMetrics(*exchanges, *pairs, *cacheSeconds)
	log.Printf("Listening on address %s", *listenAddress)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("Error starting HTTP server (%s)", err)
	}
}
