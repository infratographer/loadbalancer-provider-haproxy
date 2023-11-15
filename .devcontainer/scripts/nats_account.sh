#!/bin/bash

set -e
sudo chown -Rh vscode:vscode /workspaces/loadbalancer-provider-haproxy/.devcontainer/nsc

echo "Dumping NATS user creds file"
nsc --data-dir=/workspaces/loadbalancer-provider-haproxy/.devcontainer/nsc/nats/nsc/stores generate creds -a PROV -n USER > /tmp/user.creds

echo "Dumping NATS sys creds file"
nsc --data-dir=/workspaces/loadbalancer-provider-haproxy/.devcontainer/nsc/nats/nsc/stores generate creds -a SYS -n sys > /tmp/sys.creds
