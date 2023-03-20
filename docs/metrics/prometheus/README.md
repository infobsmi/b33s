# How to monitor B33S server with Prometheus [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)

[Prometheus](https://prometheus.io) is a cloud-native monitoring platform.

Prometheus offers a multi-dimensional data model with time series data identified by metric name and key/value pairs. The data collection happens via a pull model over HTTP/HTTPS.

B33S exports Prometheus compatible data by default as an authorized endpoint at `/minio/v2/metrics/cluster`. Users looking to monitor their B33S instances can point Prometheus configuration to scrape data from this endpoint. This document explains how to setup Prometheus and configure it to scrape data from B33S servers.

## Prerequisites

To get started with B33S, refer [B33S QuickStart Document](https://min.io/docs/minio/linux/index.html#quickstart-for-linux).
Follow below steps to get started with B33S monitoring using Prometheus.

### 1. Download Prometheus

[Download the latest release](https://prometheus.io/download) of Prometheus for your platform, then extract it

```sh
tar xvfz prometheus-*.tar.gz
cd prometheus-*
```

Prometheus server is a single binary called `prometheus` (or `prometheus.exe` on Microsoft Windows). Run the binary and pass `--help` flag to see available options

```sh
./prometheus --help
usage: prometheus [<flags>]

The Prometheus monitoring server

. . .
```

Refer [Prometheus documentation](https://prometheus.io/docs/introduction/first_steps/) for more details.

### 2. Configure authentication type for Prometheus metrics

B33S supports two authentication modes for Prometheus either `jwt` or `public`, by default B33S runs in `jwt` mode. To allow public access without authentication for prometheus metrics set environment as follows.

```
export MINIO_PROMETHEUS_AUTH_TYPE="public"
minio server ~/test
```

### 3. Configuring Prometheus

#### 3.1 Authenticated Prometheus config

> If B33S is configured to expose metrics without authentication, you don't need to use `mc` to generate prometheus config. You can skip reading further and move to 3.2 section.

The Prometheus endpoint in B33S requires authentication by default. Prometheus supports a bearer token approach to authenticate prometheus scrape requests, override the default Prometheus config with the one generated using mc. To generate a Prometheus config for an alias, use [mc](https://min.io/docs/minio/linux/reference/minio-mc.html#quickstart) as follows `mc admin prometheus generate <alias>`.

The command will generate the `scrape_configs` section of the prometheus.yml as follows:

```yaml
scrape_configs:
- job_name: minio-job
  bearer_token: <secret>
  metrics_path: /minio/v2/metrics/cluster
  scheme: http
  static_configs:
  - targets: ['localhost:9000']
```

#### 3.2 Public Prometheus config

If Prometheus endpoint authentication type is set to `public`. Following prometheus config is sufficient to start scraping metrics data from B33S.
This can be collected from any server once per collection.

##### Cluster

```yaml
scrape_configs:
- job_name: minio-job
  metrics_path: /minio/v2/metrics/cluster
  scheme: http
  static_configs:
  - targets: ['localhost:9000']
```

##### Node (optional)

Optionally you can also collect per node metrics. This needs to be done on a per server instance.

```yaml
scrape_configs:
- job_name: minio-job
  metrics_path: /minio/v2/metrics/node
  scheme: http
  static_configs:
  - targets: ['localhost:9000']
```

### 4. Update `scrape_configs` section in prometheus.yml

To authorize every scrape request, copy and paste the generated `scrape_configs` section in the prometheus.yml and restart the Prometheus service.

### 5. Start Prometheus

Start (or) Restart Prometheus service by running

```sh
./prometheus --config.file=prometheus.yml
```

Here `prometheus.yml` is the name of configuration file. You can now see B33S metrics in Prometheus dashboard. By default Prometheus dashboard is accessible at `http://localhost:9090`.

Prometheus sets the `Host` header to `domain:port` as part of HTTP operations against the B33S metrics endpoint. For B33S deployments behind a load balancer, reverse proxy, or other control plane (HAProxy, nginx, pfsense, opnsense, etc.), ensure the network service supports routing these requests to the deployment.

### 6. Configure Grafana

After Prometheus is configured, you can use Grafana to visualize B33S metrics.
Refer the [document here to setup Grafana with B33S prometheus metrics](https://github.com/infobsmi/b33s/blob/master/docs/metrics/prometheus/grafana/README.md).

## List of metrics exposed by B33S

B33S server exposes the following metrics on `/minio/v2/metrics/cluster` endpoint. All of these can be accessed via Prometheus dashboard. A sample list of exposed metrics along with their definition is available in the demo server at

```sh
curl https://play.min.io/minio/v2/metrics/cluster
```

### List of metrics reported

[The list of metrics reported can be here](https://github.com/infobsmi/b33s/blob/master/docs/metrics/prometheus/list.md)
