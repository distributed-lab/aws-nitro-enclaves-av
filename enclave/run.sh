#!/bin/bash
set -e

echo "Starting"
sleep 1

echo "Up loopback interface"
ip link set lo up || true
sleep 1

# AWS Services
echo "Start AWS IMDS egress vsock proxy"
socat TCP-LISTEN:80,bind=169.254.169.254,fork,reuseaddr,keepalive VSOCK-CONNECT:3:8001,keepalive &
echo "Start AWS KMS egress vsock proxy"
socat TCP-LISTEN:443,bind=127.0.0.2,fork,reuseaddr,keepalive VSOCK-CONNECT:3:8002,keepalive &
echo "Start AWS STS egress vsock proxy"
socat TCP-LISTEN:443,bind=127.0.0.3,fork,reuseaddr,keepalive VSOCK-CONNECT:3:8003,keepalive &

# NFS
echo "Start NFSv4 egress vsock proxy"
socat TCP-LISTEN:2049,bind=127.0.0.200,fork,reuseaddr,keepalive VSOCK-CONNECT:3:20000,keepalive &
# HTTP
echo "Start HTTP ingress vsock proxy"
socat VSOCK-LISTEN:10000,fork,keepalive TCP:127.0.0.1:8000,keepalive &

TOKEN=`curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600"`
echo "Token: $TOKEN"
AWS_REGION=`curl -H "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/placement/region`
if [[ ! "$AWS_REGION" =~ ^[a-z]{2}-[a-z]+-[0-9]+$ ]]; then
  echo "Invalid region format: $AWS_REGION"
  exit 1
fi
echo "Region: $AWS_REGION"

echo "127.0.0.2   kms.$AWS_REGION.amazonaws.com" >>/etc/hosts
echo "127.0.0.3   sts.$AWS_REGION.amazonaws.com" >>/etc/hosts

echo "Mounting persistent volume"
mkdir -p /etc/aws-nitro-enclaves-av
mount -t nfs4 127.0.0.200:/aws-nitro-enclaves-av /etc/aws-nitro-enclaves-av
sleep 1

echo "Start main process"
AWS_REGION=$AWS_REGION KV_VIPER_FILE=/etc/aws-nitro-enclaves-av/config.yaml aws-nitro-enclaves-av run service
