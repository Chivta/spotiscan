#!/bin/bash

# Registry auth
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username="${GHCR_USERNAME}" \
  --docker-password="${GHCR_PAT}" \
  --docker-email="${GHCR_EMAIL}" \
  --dry-run=client -o yaml | kubectl apply -f -

# App secrets
# Required vars:
#   DB_URL              — full postgres connection string (e.g. postgres://user:pass@postgres-svc:5432/spotiscan?sslmode=disable)
#   DB_PASSWORD         — postgres password (must match DB_URL; used by the postgres StatefulSet)
#   SPOTIFY_CLIENT_ID   — Spotify app client ID
#   SPOTIFY_CLIENT_SECRET — Spotify app client secret
#   JWT_SECRET          — random string, minimum 32 characters (e.g. openssl rand -hex 32)
kubectl create secret generic app-secrets \
  --from-literal=DB_URL="${DB_URL}" \
  --from-literal=DB_PASSWORD="${DB_PASSWORD}" \
  --from-literal=SPOTIFY_CLIENT_ID="${SPOTIFY_CLIENT_ID}" \
  --from-literal=SPOTIFY_CLIENT_SECRET="${SPOTIFY_CLIENT_SECRET}" \
  --from-literal=JWT_SECRET="${JWT_SECRET}" \
  --dry-run=client -o yaml | kubectl apply -f -