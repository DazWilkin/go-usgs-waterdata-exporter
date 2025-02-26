# Prometheus Exporter for USGS Waterdata

[![build](https://github.com/DazWilkin/go-usgs-waterdata-exporter/actions/workflows/build.yml/badge.svg)](https://github.com/DazWilkin/go-usgs-waterdata-exporter/actions/workflows/build.yml)

Uses the [Instantaneous Values (iv) service](https://waterservices.usgs.gov/docs/instantaneous-values/). There are several other [services](https://waterservices.usgs.gov/) in the portfolio.

## Table of Contents

+ [References](#references)
+ [Image](#image)
+ [Metrics](#metrics)
+ [Prometheus](#prometheus)
+ [Sigstore](#sigstore)
+ [`go tools`](#go-tools)

## References

+ [King County: Snoqualmie River at Duvall](https://flood.kingcounty.gov/gauge/32/)
+ [King County: Snoqualmie Basin](https://flood.kingcounty.gov/river/3/)
+ [USGS Waterdata: Test Tools](https://waterservices.usgs.gov/test-tools/)
+ [USGS Waterdata services](https://waterservices.usgs.gov/)
+ [USGS Waterdata: Snoqualmie River at Duvall](https://waterdata.usgs.gov/monitoring-location/12150400/#dataTypeId=continuous-00065-0&period=P365D&showMedian=false)
+ [Documentation](https://waterservices.usgs.gov/docs/instantaneous-values/instantaneous-values-details/)
+ [Snoqualmie: Duvall 2 hours](https://waterservices.usgs.gov/nwis/iv/?format=json&sites=12150400&modifiedSince=PT2H&siteStatus=all)

## Image

`ghcr.io/dazwilkin/go-usgs-waterdata-exporter:9478051dea23438044204116aa3323c80a8422f0`

## Metrics

Metrics are prefixed `usgs_waterdata_`

|Name|Type|Description|
|----|----|-----------|
|`exporter_build_info`|Counter|A metric with a constant '1' value labeled by OS version, Go version, and the Git commit of the exporter|
|`exporter_start_time`|Gauge|Exporter start time in Unix epoch seconds|
|`iv_gage_height_feet`|Gauge|Gage Height Feet|

> **NOTE** The USGS uses the spelling "Gage" instead of "Gauge"

## Run `exporter`

### Go binary

```bash
MODULE="github.com/DazWilkin/go-usgs-waterdata-exporter" # Or "."

# Sites: Snoqualmie River at Carnation, Duvall, Monroe
go run ${MODULE}/cmd/server \
--site=12149000 \
--site=12150400 \
--site=12150800
```

### Container

```bash
IMAGE="ghcr.io/dazwilkin/go-usgs-waterdata-exporter:9478051dea23438044204116aa3323c80a8422f0"

podman run \
--interactive --tty --rm \
--publish=${PORT}:${PORT} \
${IMAGE} \
--site=12149000 \
--site=12150400 \
--site=12150800
```

## Prometheus

```bash
VERS="v3.2.0"

podman run \
--interactive --tty --rm \
--name=prometheus \
--net=host \
--volume=${PWD}/prometheus.yml:/etc/prometheus/prometheus.yml \
docker.io/prom/prometheus:${VERS} \
--config.file=/etc/prometheus/prometheus.yml \
--web.enable-lifecycle
```

## Kubernetes

```bash
./kubernetes.sh
```

`pull`'s the latest image, `tag`'s it for Kubernetes local registry (`localhost:32000`), `push`'es it and then deploys using [Jsonnet](https://jsonnet.org/) (actually [`go-jsonnet`](https://github.com/google/go-jsonnet)) script ([`kubernetes.jsonnet`](./kubernetes.jsonnet)) to a cluster:

+ `Namespace`
+ `ServiceAccount`
+ `Deployment`
+ `Service`
+ `Ingress` (Tailscale)
+ `ServiceMonitor` (Prometheus Operator)
+ `VerticalPodAutoscaler`

## Sigstore

`go-usgs-waterdata-service` container images are being signed by [Sigstore](https://www.sigstore.dev/) and may be verified:

```bash
cosign verify \
--key=${PWD}/cosign.pub \
ghcr.io/dazwilkin/go-usgs-waterdata-exporter:9478051dea23438044204116aa3323c80a8422f0
```

## `go tools`

See [`go.mod`](./go.mod) `tool` section.

```bash
go tool golangci-lint run ./...
```

## JSON

See [`instantaneous_values.json`](./examples/instantaneous_values.json)