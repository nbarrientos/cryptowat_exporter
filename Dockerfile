FROM quay.io/prometheus/busybox:latest
LABEL maintainer="Nacho Barrientos <nacho@criptonita.com>"

COPY cryptowat_exporter  /bin/cryptowat_exporter

ENV CRYPTOWAT_EXCHANGES="bitstamp,kraken,coinbase-pro"
ENV CRYPTOWAT_PAIRS="btcusd,ltcusd"
ENV CRYPTOWAT_CACHESECONDS="900"
ENV WEB_LISTEN_ADDRESS=":9745"

EXPOSE      9745
ENTRYPOINT  ./bin/cryptowat_exporter --web.listen-address $WEB_LISTEN_ADDRESS --cryptowat.exchanges $CRYPTOWAT_EXCHANGES --cryptowat.pairs $CRYPTOWAT_PAIRS --cryptowat.cacheseconds $CRYPTOWAT_CACHESECONDS
