#!/bin/bash

set -euo pipefail

EIF_PATH="output/attestation-verifier.eif"
if [ ! -f "$EIF_PATH" ]; then
	echo "${EIF_PATH} not found — running ./build.base.sh"
	./build.sh
fi

docker build -t attestation-verifier:latest .
