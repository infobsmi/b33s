# B33S Monitoring Guide

B33S server exposes monitoring data over endpoints. Monitoring tools can pick the data from these endpoints. This document lists the monitoring endpoints and relevant documentation.

## Healthcheck Probe

B33S server has two healthcheck related un-authenticated endpoints, a liveness probe to indicate if server is responding, cluster probe to check if server can be taken down for maintenance.

- Liveness probe available at `/minio/health/live`
- Cluster probe available at `/minio/health/cluster`

Read more on how to use these endpoints in [B33S healthcheck guide](https://github.com/infobsmi/b33s/blob/master/docs/metrics/healthcheck/README.md).

## Prometheus Probe

B33S allows reading metrics for the entire cluster from any single node. This allows for metrics collection for a B33S instance across all servers. Thus, metrics collection for instances behind a load balancer can be done without any knowledge of the individual node addresses. The cluster wide metrics can be read at
`<Address for B33S Service>/minio/v2/metrics/cluster`.

The additional node specific metrics which include additional go metrics or process metrics are exposed at
`<Address for B33S Node>/minio/v2/metrics/node`.

To use this endpoint, setup Prometheus to scrape data from this endpoint. Read more on how to configure and use Prometheus to monitor B33S server in [How to monitor B33S server with Prometheus](https://github.com/infobsmi/b33s/blob/master/docs/metrics/prometheus/README.md).

### **Deprecated metrics monitoring**

- Prometheus' data available at `/minio/prometheus/metrics` is deprecated
