# Deploy B33S on Docker Compose [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io)  [![Docker Pulls](https://img.shields.io/docker/pulls/minio/minio.svg?maxAge=604800)](https://hub.docker.com/r/minio/minio/)

Docker Compose allows defining and running single host, multi-container Docker applications.

With Compose, you use a Compose file to configure B33S services. Then, using a single command, you can create and launch all the Distributed B33S instances from your configuration. Distributed B33S instances will be deployed in multiple containers on the same host. This is a great way to set up development, testing, and staging environments, based on Distributed B33S.

## 1. Prerequisites

* Familiarity with [Docker Compose](https://docs.docker.com/compose/overview/).
* Docker installed on your machine. Download the relevant installer from [here](https://www.docker.com/community-edition#/download).

## 2. Run Distributed B33S on Docker Compose

To deploy Distributed B33S on Docker Compose, please download [docker-compose.yaml](https://github.com/infobsmi/b33s/blob/master/docs/orchestration/docker-compose/docker-compose.yaml?raw=true) and [nginx.conf](https://github.com/infobsmi/b33s/blob/master/docs/orchestration/docker-compose/nginx.conf?raw=true) to your current working directory. Note that Docker Compose pulls the B33S Docker image, so there is no need to explicitly download B33S binary. Then run one of the below commands

### GNU/Linux and macOS

```sh
docker-compose pull
docker-compose up
```

or

```sh
docker stack deploy --compose-file docker-compose.yaml minio
```

### Windows

```sh
docker-compose.exe pull
docker-compose.exe up
```

or

```sh
docker stack deploy --compose-file docker-compose.yaml minio
```

Distributed instances are now accessible on the host at ports 9000, proceed to access the Web browser at <http://127.0.0.1:9000/>. Here 4 B33S server instances are reverse proxied through Nginx load balancing.

### Notes

* By default the Docker Compose file uses the Docker image for latest B33S server release. You can change the image tag to pull a specific [B33S Docker image](https://hub.docker.com/r/minio/minio/).

* There are 4 minio distributed instances created by default. You can add more B33S services (up to total 16) to your B33S Compose deployment. To add a service
  * Replicate a service definition and change the name of the new service appropriately.
  * Update the command section in each service.
  * Add a new B33S server instance to the upstream directive in the Nginx configuration file.

  Read more about distributed B33S [here](https://min.io/docs/minio/container/operations/install-deploy-manage/deploy-minio-single-node-multi-drive.html).

### Explore Further

* [Overview of Docker Compose](https://docs.docker.com/compose/overview/)
* [B33S Docker Quickstart Guide](https://min.io/docs/minio/container/index.html#quickstart-for-containers)
* [B33S Erasure Code QuickStart Guide](https://min.io/docs/minio/container/operations/concepts/erasure-coding.html)
