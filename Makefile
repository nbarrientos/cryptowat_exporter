DOCKER_ARCHS ?= amd64 armv7 arm64 ppc64le

include Makefile.common

STATICCHECK_IGNORE =

DOCKER_IMAGE_NAME ?= cryptowat_exporter

ifdef DEBUG
	bindata_flags = -debug
endif