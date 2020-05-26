ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="Nacho Barrientos <nacho@criptonita.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/prometheus-cryptowat-exporter  /bin/prometheus-cryptowat-exporter

ENV CRYPTOWAT_EXCHANGES="bitstamp,kraken,coinbase-pro"
ENV CRYPTOWAT_PAIRS="btcusd,ltcusd"
ENV WEB_LISTEN_ADDRESS=":9150"

EXPOSE      9150
ENTRYPOINT  ./bin/prometheus-cryptowat-exporter --web.listen-address $WEB_LISTEN_ADDRESS --cryptowat.exchanges $CRYPTOWAT_EXCHANGES --cryptowat.pairs $CRYPTOWAT_PAIRS