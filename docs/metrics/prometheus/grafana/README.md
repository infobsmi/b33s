# How to monitor B33S server with Grafana [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)

[Grafana](https://grafana.com/) allows you to query, visualize, alert on and understand your metrics no matter where they are stored. Create, explore, and share dashboards with your team and foster a data driven culture.

## Prerequisites

- Prometheus and B33S configured as explained in [document here](https://github.com/infobsmi/b33s/blob/master/docs/metrics/prometheus/README.md).
- Grafana installed as explained [here](https://grafana.com/grafana/download).

## B33S Grafana Dashboard

Visualize B33S metrics with our official Grafana dashboard available on the [Grafana dashboard portal](https://grafana.com/grafana/dashboards/13502).

Refer to the dashboard [json file here](https://raw.githubusercontent.com/minio/minio/master/docs/metrics/prometheus/grafana/minio-dashboard.json).

![Grafana](https://raw.githubusercontent.com/minio/minio/master/docs/metrics/prometheus/grafana/grafana-minio.png)

Replication metrics can also be viewed in the Grafana dashboard using [json file here](https://raw.githubusercontent.com/minio/minio/master/docs/metrics/prometheus/grafana/minio-replication.json).

![Grafana](https://raw.githubusercontent.com/minio/minio/master/docs/metrics/prometheus/grafana/grafana-replication.png)
