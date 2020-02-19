#!/usr/bin/env bash

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)

echo '>>> Loading image in Kind'
kind load docker-image yyklll/skeleton:latest

echo '>>> Installing'
helm upgrade -i skeleton ${REPO_ROOT}/charts/skeleton --namespace=default
kubectl set image deployment/skeleton skeleton=yyklll/skeleton:latest
kubectl rollout status deployment/skeleton
