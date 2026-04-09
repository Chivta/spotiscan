#!/bin/bash

# All secrets are read from .env.secrets (not version-controlled).
# Required keys:
#   GHCR_USERNAME       — GitHub username for container registry
#   GHCR_PAT            — GitHub personal access token (read:packages scope)
#   GHCR_EMAIL          — GitHub account email
#   DB_URL              — full postgres connection string (e.g. postgres://user:pass@postgres-svc:5432/ruscan?sslmode=disable)
#   DB_PASSWORD         — postgres password (must match DB_URL; used by the postgres StatefulSet)
#   SPOTIFY_CLIENT_ID   — Spotify app client ID
#   SPOTIFY_CLIENT_SECRET — Spotify app client secret
#   JWT_SECRET          — random string, minimum 32 characters (e.g. openssl rand -hex 32)
#   LastFMAPIKey        —  Last.fm API key

set -o allexport
source .env.secrets
set +o allexport

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