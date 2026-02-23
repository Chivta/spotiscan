#!/bin/bash

# Registry auth
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username="${GHCR_USERNAME}" \
  --docker-password="${GHCR_PAT}" \
  --docker-email="${GHCR_EMAIL}" \
  --dry-run=client -o yaml | kubectl apply -f -

# App secrets
kubectl create secret generic app-secrets \
  --from-env-file=.env.secrets \
  --dry-run=client -o yaml | kubectl apply -f -