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

			sleepSeconds, _ := strconv.ParseInt(cacheSeconds, 10, 32)
			log.Printf("Sleeping for %d seconds", sleepSeconds)
			time.Sleep(time.Duration(sleepSeconds) * time.Second)
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
)

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9150").String()
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
