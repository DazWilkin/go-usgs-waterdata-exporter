#!/usr/bin/env bash

# Updated by GitHub Workflow actions
OLD_IMAGE="ghcr.io/dazwilkin/go-usgs-waterdata-exporter:cb507e76390674dab723b868bb0b1df8d76bc23d"

# Replace "ghcr.io/dazwilkin" with "localhost:32000"
NEW_IMAGE="localhost:32000/${OLD_IMAGE#ghcr.io/dazwilkin/}"

podman pull ${OLD_IMAGE}

podman tag \
  "${OLD_IMAGE}" \
  "${NEW_IMAGE}"

podman push "${NEW_IMAGE}"

CONFIG="${PWD}/tmp/kubernetes.$(date +%y%m%d).yaml"

# Generate and persist the config for auditability
jsonnet \
--ext-str image="${NEW_IMAGE}" \
./kubernetes.jsonnet > "${CONFIG}"

# Apply the persisted config
kubectl apply --filename="${CONFIG}"
