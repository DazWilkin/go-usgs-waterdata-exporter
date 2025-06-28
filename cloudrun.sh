#!/usr/bin/env bash

if [ -z "${PROJECT}" ]
then
  echo "Expected 'PROJECT' to be set"
  exit 1
fi

if [ -z "${NAME}" ]
then
  echo "Expected 'NAME' to be set"
  exit 1
fi

if [ -z "${REGION}" ]
then
  echo "Expected 'REGION' to be set"
  exit 1
fi

if [ -z "${REPOSITORY}" ]
then
  echo "Expected 'REPOSITORY' to be set"
  exit 1
fi

OLD_IMAGE="ghcr.io/dazwilkin/go-usgs-waterdata-exporter:7423f4228c58abb3b2f365ab0930af275c626cd2"
NEW_IMAGE="${REGION}-docker.pkg.dev/${PROJECT}/${REPOSITORY}/${OLD_IMAGE#ghcr.io/dazwilkin/}"

podman pull ${OLD_IMAGE}

podman tag \
  "${OLD_IMAGE}" \
  "${NEW_IMAGE}"

gcloud auth print-access-token |\
podman login "${REGION}-docker.pkg.dev" \
--username=oauth2accesstoken \
--password-stdin

podman push "${NEW_IMAGE}"

# Generate and persist the config for auditability
CONFIG="${PWD}/tmp/cloudrun.$(date +%y%m%d).json"

jsonnet \
--ext-str image="${IMAGE}" \
./cloudrun.jsonnet > "${CONFIG}"

# Apply the persisted config
gcloud run deploy ${NAME} \
--image=${NEW_IMAGE} \
--no-allow-unauthenticated \
--max-instances=1 \
--region=${REGION} \
--args="--endpoint=:8080","--path=/metrics","--sitecode=12149000","--sitecode=12150400","--sitecode=12150800" \
--project=${PROJECT}
