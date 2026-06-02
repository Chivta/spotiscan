# Ruscan — Metrics Reference

Complete reference of all Prometheus metrics exposed or pushed by the ruscan stack, intended for Grafana dashboard design.

---

## Collection Overview

| Source | Mechanism | Namespace | Interval |
|---|---|---|---|
| `api` | ServiceMonitor scrape `:8080/metrics` | `ruscan` | 15s |
| `scan-worker` | ServiceMonitor scrape `:8080/metrics` | `ruscan` | 15s |
| `spotify-gateway` | ServiceMonitor scrape `:8080/metrics` | `ruscan` | 15s |
| `scraper` | Pushgateway push (`job="scraper"`) | `monitoring` | on-run |
| Redis | `prometheus-redis-exporter` ServiceMonitor | `monitoring` | default |
| Postgres | `prometheus-postgres-exporter` ServiceMonitor | `monitoring` | 15s |
| RabbitMQ | `prometheus-rabbitmq-exporter` ServiceMonitor | `monitoring` | 15s |
| Nodes | `node-exporter` (kube-prometheus-stack) | `monitoring` | default |
| Kubernetes | `kube-state-metrics` (kube-prometheus-stack) | `monitoring` | default |

Prometheus is configured with `serviceMonitorSelectorNilUsesHelmValues: false` and empty selectors, so it scrapes **all** ServiceMonitors across all namespaces. Retention is **7 days**.

---

## Application Metrics

### api

> `internal/api/metrics.go` — gin middleware + promhttp handler

| Metric | Type | Labels | Description |
|---|---|---|---|
| `http_requests_total` | Counter | `method`, `path`, `status` | Total HTTP requests received |
| `http_request_duration_seconds` | Histogram | `method`, `path` | Request latency (DefBuckets) |
| `http_requests_in_flight` | Gauge | — | Concurrent requests being processed |

**Label values:**
- `method`: HTTP verb (`GET`, `POST`, …)
- `path`: gin full path pattern (e.g. `/api/v1/artists/:id`)
- `status`: HTTP status code string (`200`, `404`, …)

**Shared metrics also present on this target:**

| Metric | Type | Labels | Description |
|---|---|---|---|
| `redis_latency_seconds` | Histogram | `operation` | Redis command latency (DefBuckets) |
| `postgres_latency_seconds` | Histogram | `operation` | Postgres query latency via pgx tracer (DefBuckets) |
| `errors_total` | Counter | `component` | Error log events from zerolog hook |

---

### scan-worker

> `internal/scanner/metrics.go` — promhttp handler only; shared metrics wired in

| Metric | Type | Labels | Description |
|---|---|---|---|
| `redis_latency_seconds` | Histogram | `operation` | Redis command latency (DefBuckets) |
| `postgres_latency_seconds` | Histogram | `operation` | Postgres query latency via pgx tracer (DefBuckets) |
| `errors_total` | Counter | `component` | Error log events from zerolog hook |

**`postgres_latency_seconds` operation values:** `SELECT`, `INSERT`, `UPDATE`, `DELETE`, `error`, `unknown` (taken from pgx CommandTag).

**`redis_latency_seconds` operation values:** individual command names (`get`, `set`, `hget`, …) + `pipeline`.

---

### spotify-gateway

> `internal/spotify/metrics.go` — promhttp handler + Spotify client instrumentation

| Metric | Type | Labels | Description |
|---|---|---|---|
| `spotify_api_duration_seconds` | Histogram | `method` | Duration of outbound Spotify API calls (DefBuckets) |
| `errors_total` | Counter | `component` | Error log events from zerolog hook |

**`method` label:** identifies the Spotify API method called (e.g. `GetArtist`, `Search`, …).

---

### scraper (Pushgateway)

> `internal/scraper/metrics.go` — pushed to Pushgateway with `job="scraper"`

| Metric | Type | Labels | Description |
|---|---|---|---|
| `scraper_ru_artists_total` | Gauge | — | Current count of Russian artists in the database |

The scraper pushes the **full `DefaultGatherer`**, so standard Go runtime metrics (`go_goroutines`, `go_memstats_*`, etc.) are also present under `job="scraper"`.

**PromQL note:** because this is pushed, use `scraper_ru_artists_total` without rate/increase — it is a snapshot gauge.

---

## Shared / Cross-Cutting Metrics

These are registered once in `internal/shared/metrics/` and appear on every service target.

### `errors_total`

```
errors_total{component="<name>"}
```

Incremented by the zerolog `MetricsHook` on every `Error`-level log event. Component name is set per-service at logger initialisation (e.g. `"api"`, `"scan-worker"`, `"spotify-gateway"`).

**Useful query:**
```promql
sum by (component) (increase(errors_total[5m]))
```

### `redis_latency_seconds`

```
redis_latency_seconds{operation="<cmd>"}
```

Instrumented via a `redis.Hook`. Covers both individual commands and `pipeline` batches.

**Useful queries:**
```promql
# P99 latency per operation
histogram_quantile(0.99, sum by (le, operation) (rate(redis_latency_seconds_bucket[5m])))

# Error rate (operation="error" is not currently produced by the hook, but watch for it)
```

### `postgres_latency_seconds`

```
postgres_latency_seconds{operation="<cmd>"}
```

Instrumented via a `pgx.QueryTracer`. Operation is the first word of the CommandTag (`SELECT`, `INSERT`, etc.) or `"error"` on failure.

**Useful queries:**
```promql
histogram_quantile(0.95, sum by (le, operation) (rate(postgres_latency_seconds_bucket[5m])))
```

---

## Infrastructure Exporters

### Redis — `prometheus-redis-exporter`

Target: `redis://redis.ruscan.svc.cluster.local:6379` (instance label `name="ruscan"`)

Key metrics exposed by the exporter:

| Metric | Type | Description |
|---|---|---|
| `redis_connected_clients` | Gauge | Number of client connections |
| `redis_blocked_clients` | Gauge | Clients waiting on blocking command |
| `redis_used_memory_bytes` | Gauge | Total memory allocated by Redis |
| `redis_used_memory_rss_bytes` | Gauge | RSS memory from OS perspective |
| `redis_keyspace_hits_total` | Counter | Cache hits |
| `redis_keyspace_misses_total` | Counter | Cache misses |
| `redis_commands_processed_total` | Counter | Total commands processed |
| `redis_evicted_keys_total` | Counter | Keys evicted due to maxmemory |
| `redis_expired_keys_total` | Counter | Keys expired |
| `redis_db_keys` | Gauge | Total keys per DB (`db` label) |
| `redis_uptime_in_seconds` | Gauge | Redis uptime |

**Hit rate:**
```promql
rate(redis_keyspace_hits_total[5m]) /
(rate(redis_keyspace_hits_total[5m]) + rate(redis_keyspace_misses_total[5m]))
```

---

### Postgres — `prometheus-postgres-exporter`

Target: `postgres.ruscan.svc.cluster.local:5432`, database `ruscan`

Key metrics:

| Metric | Type | Labels | Description |
|---|---|---|---|
| `pg_up` | Gauge | — | 1 if connection succeeds |
| `pg_stat_activity_count` | Gauge | `datname`, `state` | Active connections by state |
| `pg_stat_database_tup_fetched_total` | Counter | `datname` | Rows fetched |
| `pg_stat_database_tup_inserted_total` | Counter | `datname` | Rows inserted |
| `pg_stat_database_tup_updated_total` | Counter | `datname` | Rows updated |
| `pg_stat_database_tup_deleted_total` | Counter | `datname` | Rows deleted |
| `pg_stat_database_blks_hit_total` | Counter | `datname` | Buffer cache hits |
| `pg_stat_database_blks_read_total` | Counter | `datname` | Disk block reads |
| `pg_stat_database_deadlocks_total` | Counter | `datname` | Deadlocks detected |
| `pg_stat_database_conflicts_total` | Counter | `datname` | Conflicts with recovery |
| `pg_locks_count` | Gauge | `datname`, `mode` | Lock counts by mode |
| `pg_postmaster_start_time_seconds` | Gauge | — | Postgres start time (restart detection) |

**Cache hit ratio:**
```promql
rate(pg_stat_database_blks_hit_total{datname="ruscan"}[5m]) /
(rate(pg_stat_database_blks_hit_total{datname="ruscan"}[5m]) + rate(pg_stat_database_blks_read_total{datname="ruscan"}[5m]))
```

---

### RabbitMQ — `prometheus-rabbitmq-exporter`

Target: `http://rabbitmq.ruscan.svc.cluster.local:15672` (management API)

Key metrics:

| Metric | Type | Labels | Description |
|---|---|---|---|
| `rabbitmq_up` | Gauge | — | 1 if management API reachable |
| `rabbitmq_queue_messages` | Gauge | `queue`, `vhost` | Messages ready + unacked |
| `rabbitmq_queue_messages_ready` | Gauge | `queue`, `vhost` | Messages ready to deliver |
| `rabbitmq_queue_messages_unacknowledged` | Gauge | `queue`, `vhost` | Messages delivered but unacked |
| `rabbitmq_queue_consumers` | Gauge | `queue`, `vhost` | Consumer count per queue |
| `rabbitmq_queue_message_bytes` | Gauge | `queue`, `vhost` | Bytes in queue |
| `rabbitmq_connections` | Gauge | — | Total connections |
| `rabbitmq_channels` | Gauge | — | Total channels |
| `rabbitmq_consumers` | Gauge | — | Total consumers |
| `rabbitmq_exchanges` | Gauge | — | Total exchanges |
| `rabbitmq_node_mem_used` | Gauge | `node` | Memory used by node |
| `rabbitmq_node_mem_limit` | Gauge | `node` | Memory limit |
| `rabbitmq_node_disk_free` | Gauge | `node` | Free disk space |
| `rabbitmq_node_fd_used` | Gauge | `node` | File descriptors used |

---

## Standard Platform Metrics (kube-prometheus-stack)

### node-exporter

Available on all nodes. Key metrics for dashboards:

| Metric | Description |
|---|---|
| `node_cpu_seconds_total{mode}` | CPU time by mode (use `rate` + `1 - idle`) |
| `node_memory_MemAvailable_bytes` | Available memory |
| `node_memory_MemTotal_bytes` | Total memory |
| `node_filesystem_avail_bytes` | Filesystem free space |
| `node_network_receive_bytes_total` | Network RX |
| `node_network_transmit_bytes_total` | Network TX |
| `node_load1` / `node_load5` / `node_load15` | Load averages |

### kube-state-metrics

Key metrics for pod/deployment health:

| Metric | Description |
|---|---|
| `kube_deployment_status_replicas_available` | Ready replicas per deployment |
| `kube_deployment_spec_replicas` | Desired replicas |
| `kube_pod_status_phase` | Pod phase counts |
| `kube_pod_container_status_restarts_total` | Container restarts |
| `kube_pod_container_resource_requests` | CPU/memory requests |
| `kube_pod_container_resource_limits` | CPU/memory limits |

---

## Suggested Grafana Dashboard Structure

### Row 1 — API Health
- Request rate: `sum(rate(http_requests_total[5m])) by (path)`
- Error rate (5xx): `sum(rate(http_requests_total{status=~"5.."}[5m]))`
- P95 latency: `histogram_quantile(0.95, sum by (le, path) (rate(http_request_duration_seconds_bucket[5m])))`
- In-flight requests: `http_requests_in_flight`

### Row 2 — Errors
- Error rate by component: `sum by (component) (rate(errors_total[5m]))`
- Stat panel: total errors in last hour per service

### Row 3 — Database
- Postgres cache hit ratio
- Postgres connections by state
- Postgres query latency P99: `histogram_quantile(0.99, sum by (le, operation) (rate(postgres_latency_seconds_bucket[5m])))`
- Deadlocks: `rate(pg_stat_database_deadlocks_total{datname="ruscan"}[5m])`

### Row 4 — Redis
- Hit rate
- Memory used vs available
- Redis latency P99: `histogram_quantile(0.99, sum by (le, operation) (rate(redis_latency_seconds_bucket[5m])))`
- Evictions: `rate(redis_evicted_keys_total[5m])`

### Row 5 — RabbitMQ / Queue
- Messages ready per queue
- Unacknowledged messages per queue
- Consumer count per queue
- Node memory usage

### Row 6 — Spotify Gateway
- Spotify API call rate by method: `sum by (method) (rate(spotify_api_duration_seconds_count[5m]))`
- Spotify API P95 latency: `histogram_quantile(0.95, sum by (le, method) (rate(spotify_api_duration_seconds_bucket[5m])))`

### Row 7 — Scraper
- Russian artists total: `scraper_ru_artists_total`

### Row 8 — Infrastructure
- Pod restart rate by deployment
- CPU usage per node
- Memory pressure per node
