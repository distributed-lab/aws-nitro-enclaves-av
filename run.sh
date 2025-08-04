#!/bin/bash
set -e

echo "Starting"
sleep 5

echo "Setup /etc/hosts"
echo "127.0.0.2   kms.us-east-1.amazonaws.com kms.us-east-2.amazonaws.com kms.us-west-1.amazonaws.com kms.us-west-2.amazonaws.com kms.ap-south-1.amazonaws.com kms.ap-northeast-1.amazonaws.com kms.ap-northeast-2.amazonaws.com kms.ap-northeast-3.amazonaws.com kms.ap-southeast-1.amazonaws.com kms.ap-southeast-2.amazonaws.com kms.ca-central-1.amazonaws.com kms.eu-central-1.amazonaws.com kms.eu-west-1.amazonaws.com kms.eu-west-2.amazonaws.com kms.eu-west-3.amazonaws.com kms.eu-north-1.amazonaws.com kms.sa-east-1.amazonaws.com" >>/etc/hosts
echo "127.0.0.3   sts.us-east-1.amazonaws.com sts.us-east-2.amazonaws.com sts.us-west-1.amazonaws.com sts.us-west-2.amazonaws.com sts.ap-south-1.amazonaws.com sts.ap-northeast-1.amazonaws.com sts.ap-northeast-2.amazonaws.com sts.ap-northeast-3.amazonaws.com sts.ap-southeast-1.amazonaws.com sts.ap-southeast-2.amazonaws.com sts.ca-central-1.amazonaws.com sts.eu-central-1.amazonaws.com sts.eu-west-1.amazonaws.com sts.eu-west-2.amazonaws.com sts.eu-west-3.amazonaws.com sts.eu-north-1.amazonaws.com sts.sa-east-1.amazonaws.com" >>/etc/hosts

# Forwarding rules
# '->' forward traffic from Enclave to EC2
# Usage: IP:INET_PORT -> CID:VSOCK_PORT
# '<-' forward traffic from EC2 to Enclave
# Usage: INET_PORT <- VSOCK_PORT
rules=(
    "127.0.0.2:443 -> 3:8002"
    "127.0.0.3:443 -> 3:8003"
    "127.0.0.200:2049 -> 3:20000"
    "169.254.169.254:80 -> 3:16900"
    "8000 <- 8080"
)

echo "Up loopback interface"
ip link set lo up || true
sleep 5

for rule in "${rules[@]}"; do
  if [[ "$rule" == *"->"* ]]; then
    left=$(echo "${rule%%->*}" | awk '{$1=$1}1')
    right=$(echo "${rule##*->}" | awk '{$1=$1}1')

    inet_ip="${left%%:*}"
    inet_port="${left##*:}"
    vsock_cid="${right%%:*}"
    vsock_port="${right##*:}"

    echo "Assign $inet_ip to lo"
    if ! ip addr show dev lo | grep -q "$inet_ip"; then
      ip addr add "$inet_ip/32" dev lo:0
      ip link set dev lo:0 up
    fi
    sleep 1

    echo "Start $rule socat proxy"
    socat TCP-LISTEN:$inet_port,bind=$inet_ip,fork,reuseaddr,keepalive VSOCK-CONNECT:$vsock_cid:$vsock_port,keepalive &
  elif [[ "$rule" == *"<-"* ]]; then
    inet_port=$(echo "${rule%%<-*}" | awk '{$1=$1}1')
    vsock_port=$(echo "${rule##*<-}" | awk '{$1=$1}1')

    echo "Start $rule socat proxy"
    socat VSOCK-LISTEN:$vsock_port,fork,keepalive TCP:127.0.0.1:$inet_port,keepalive &
  fi
done
sleep 5

# TOOD: Add validation
TOKEN=`curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600"`
REGION=`curl -H "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/placement/region`

echo "127.0.0.2   kms.$REGION.amazonaws.com" >>/etc/hosts
echo "127.0.0.3   sts.$REGION.amazonaws.com" >>/etc/hosts

echo "Mounting persistent volume to /shared"
mkdir -p /shared
mount -t nfs4 127.0.0.200:/ /shared
sleep 5

echo "Start main process"
KV_VIPER_FILE=/shared/config.yaml aws-nitro-enclaves-av run service
