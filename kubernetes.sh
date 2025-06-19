#!/usr/bin/env bash

# Updated by GitHub Workflow actions
OLD_IMAGE="ghcr.io/dazwilkin/go-usgs-waterdata-exporter:ef6d89634b11900b2082d53406eab8f2530e7979"

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
