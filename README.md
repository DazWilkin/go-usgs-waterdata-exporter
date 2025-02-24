# Prometheus Exporter for USGS Waterdata

## Table of Contents

+ [References](#references)
+ [Prometheus](#prometheus)
+ [Podman](#podman)
+ [`go tools`](#go-tools)

## References

+ [King County: Snoqualmie River at Duvall](https://flood.kingcounty.gov/gauge/32/)
+ [King County: Snoqualmie Basin](https://flood.kingcounty.gov/river/3/)
+ [USGS Waterdata: Test Tools](https://waterservices.usgs.gov/test-tools/)
+ [USGS Waterdata services](https://waterservices.usgs.gov/)
+ [USGS Waterdata: Snoqualmie River at Duvall](https://waterdata.usgs.gov/monitoring-location/12150400/#dataTypeId=continuous-00065-0&period=P365D&showMedian=false)
+ [Documentation](https://waterservices.usgs.gov/docs/instantaneous-values/instantaneous-values-details/)
+ [Snoqualmie: Duvall 2 hours](https://waterservices.usgs.gov/nwis/iv/?format=json&sites=12150400&modifiedSince=PT2H&siteStatus=all)

## Run

```bash
MODULE="github.com/DazWilkin/go-usgs-waterdata-exporter" # Or "."

# Sites: Snoqualmie River at Carnation, Duvall, Monroe
go run ${MODULE}/cmd/server \
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

## Podman

Can't local build with podman because [`Dockerfile`](./Dockerfile) uses `--platform=${TARGETARCH}` for multi-platform builds.

```bash
podman build \
--tag=go-usgs-waterdata-exporter \
--file=./Dockerfile \
--build-arg=COMMIT=${COMMIT} \
--build-arg=VERSION="v0.0.1" \
--build-arg=TARGETOS=linux \
--build-arg=TARGETARCH=amd64 \
${PWD}
```
```
[1/2] STEP 1/13: FROM docker.io/golang:1.24.0 AS build
Error: unable to parse platform "amd64": invalid platform syntax for "amd64" (use OS/ARCH[/VARIANT][,...])
```
```bash
podman build \
--platform=linux/amd64 \
--tag=go-usgs-waterdata-exporter \
--file=./Dockerfile \
--build-arg=COMMIT=${COMMIT} \
--build-arg=VERSION="v0.0.1" \
${PWD}
```
```
[1/2] STEP 1/13: FROM docker.io/golang:1.24.0 AS build
Error: unable to parse platform "amd64": invalid platform syntax for "amd64" (use OS/ARCH[/VARIANT][,...])
```

## `go tools`

See [`go.mod`](./go.mod) `tool` section.

```bash
go tool golangci-lint run ./...
```

## JSON

See [`instantaneous_values.json`](./examples/instantaneous_values.json)