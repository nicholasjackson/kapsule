#! /bin/bash

# Enable transit secrets engine
vault secrets enable transit

# Create a new encryption key
vault write transit/keys/kapsule exportable=true type=rsa-4096