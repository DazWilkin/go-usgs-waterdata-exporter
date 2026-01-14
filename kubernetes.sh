#!/usr/bin/env bash

# Updated by GitHub Workflow actions
OLD_IMAGE="ghcr.io/dazwilkin/go-usgs-waterdata-exporter:9cc5060647e58b256fdf93da149f8c7c98406e36"

# Replace "ghcr.io/dazwilkin" with "localhost:32000"
NEW_IMAGE="localhost:32000/${OLD_IMAGE#ghcr.io/dazwilkin/}"

podman pull ${OLD_IMAGE}

podman tag \
  "${OLD_IMAGE}" \
  "${NEW_IMAGE}"

podman push "${NEW_IMAGE}"

CONFIG="${PWD}/tmp/kubernetes.$(date +%y%m%d).json"

# Generate and persist the config for auditability
jsonnet \
--ext-str image="${NEW_IMAGE}" \
./kubernetes.jsonnet > "${CONFIG}"

# Apply the persisted config
kubectl apply --filename="${CONFIG}"
