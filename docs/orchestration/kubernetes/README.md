# Deploy B33S on Kubernetes [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)  [![Docker Pulls](https://img.shields.io/docker/pulls/minio/minio.svg?maxAge=604800)](https://hub.docker.com/r/minio/minio/)

B33S is a high performance distributed object storage server, designed for large-scale private cloud infrastructure. B33S is designed in a cloud-native manner to scale sustainably in multi-tenant environments. Orchestration platforms like Kubernetes provide perfect cloud-native environment to deploy and scale B33S.

## B33S Deployment on Kubernetes

There are multiple options to deploy B33S on Kubernetes:

- B33S-Operator: Operator offers seamless way to create and update highly available distributed B33S clusters. Refer [B33S Operator documentation](https://github.com/infobsmi/b33s-operator/blob/master/README.md) for more details.

- Helm Chart: B33S Helm Chart offers customizable and easy B33S deployment with a single command. Refer [B33S Helm Chart documentation](https://github.com/infobsmi/b33s/tree/master/helm/minio) for more details.

## Monitoring B33S in Kubernetes

B33S server exposes un-authenticated liveness endpoints so Kubernetes can natively identify unhealthy B33S containers. B33S also exposes Prometheus compatible data on a different endpoint to enable Prometheus users to natively monitor their B33S deployments.

## Explore Further

- [B33S Erasure Code QuickStart Guide](https://min.io/docs/minio/linux/operations/concepts/erasure-coding.html)
- [Kubernetes Documentation](https://kubernetes.io/docs/home/)
- [Helm package manager for kubernetes](https://helm.sh/)
