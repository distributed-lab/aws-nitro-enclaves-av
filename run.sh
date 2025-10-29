#!/bin/bash -e

# Start vsock proxies
# Outbound enclave connections

# IMDS
socat VSOCK-LISTEN:8001,fork,keepalive TCP:169.254.169.254:80,keepalive &
# AWS KMS
socat VSOCK-LISTEN:8002,fork,keepalive TCP:kms.$AWS_REGION.amazonaws.com:443,keepalive &
# AWS STS
socat VSOCK-LISTEN:8003,fork,keepalive TCP:sts.$AWS_REGION.amazonaws.com:443,keepalive &

# NFS Server
socat VSOCK-LISTEN:20000,fork,keepalive TCP:$NFS_SERVER,keepalive &


# Inbound enclave connections

# HTTP
socat TCP-LISTEN:$HTTP_PORT,fork,reuseaddr,keepalive VSOCK-CONNECT:$ENCLAVE_CID:10000,keepalive &

nitro-cli run-enclave --eif-path /root/attestation-verifier.eif --enclave-cid $ENCLAVE_CID --cpu-count $CPU_COUNT --memory $MEMORY_MIB $EXTRA_OPTIONS
enclave_id=$(nitro-cli describe-enclaves | jq -r ".[0].EnclaveID")
echo "-------------------------------"
echo "Enclave ID is $enclave_id"
echo "-------------------------------"

sed -i "s/ENCLAVE_ID_TO_REPLACE/$enclave_id/g" /opt/prestop-hook.sh

nitro-cli console --enclave-id $enclave_id || true && tail -f /dev/null
