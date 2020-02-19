#!/usr/bin/env bash

set -o errexit

function finish {
  echo '>>> Test logs'
  kubectl logs -l app=skeleton
}
trap finish EXIT

echo '>>> Start End-to-End tests'
helm test skeleton
