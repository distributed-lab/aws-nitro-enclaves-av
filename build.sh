#!/bin/bash

set -euo pipefail

docker build -t attestation-verifier-base:latest --target attestation-verifier-base .
mkdir -p output
nitro-cli build-enclave --docker-uri attestation-verifier-base:latest --output-file output/attestation-verifier.eif
