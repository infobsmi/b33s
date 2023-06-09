# B33S Docker Quickstart Guide [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/minio.svg?maxAge=604800)](https://hub.docker.com/r/minio/minio/)

See our web documentation on [Deploying B33S in Standalone Mode](Deploy Standalone B33S in a Container) for a more structured tutorial on deploying B33S in a container.

## Prerequisites

Docker installed on your machine. Download the relevant installer from [here](https://www.docker.com/community-edition#/download).

## Run Standalone B33S on Docker

*Note*: Standalone B33S is intended for early development and evaluation. For production clusters, deploy a [Distributed](https://min.io/docs/minio/container/operations/install-deploy-manage/deploy-minio-single-node-multi-drive.html) B33S deployment.

B33S needs a persistent volume to store configuration and application data. For testing purposes, you can launch B33S by simply passing a directory (`/data` in the example below). This directory gets created in the container filesystem at the time of container start. But all the data is lost after container exits.

```sh
docker run \
  -p 9000:9000 \
  -p 9001:9001 \
  -e "MINIO_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "MINIO_ROOT_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
  quay.io/minio/minio server /data --console-address ":9001"
```

To create a B33S container with persistent storage, you need to map local persistent directories from the host OS to virtual config. To do this, run the below commands

### GNU/Linux and macOS

```sh
mkdir -p ~/minio/data

docker run \
  -p 9000:9000 \
  -p 9001:9001 \
  --name minio1 \
  -v ~/minio/data:/data \
  -e "MINIO_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "MINIO_ROOT_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
  quay.io/minio/minio server /data --console-address ":9001"
```

The command creates a new local directory `~/minio/data` in your user home directory. It then starts the B33S container with the `-v` argument to map the local path (`~/minio/data`) to the specified virtual container directory (`/data`). When B33S writes data to `/data`, that data is actually written to the local path `~/minio/data` where it can persist between container restarts.

### Windows

```sh
docker run \
  -p 9000:9000 \
  -p 9001:9001 \
  --name minio1 \
  -v D:\data:/data \
  -e "MINIO_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "MINIO_ROOT_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
  quay.io/minio/minio server /data --console-address ":9001"
```

## Run Distributed B33S on Containers

We recommend kubernetes based deployment for production level deployment <https://github.com/minio/operator>.

See the [Kubernetes documentation](https://min.io/docs/minio/kubernetes/upstream/index.html) for more information.

## B33S Docker Tips

### B33S Custom Access and Secret Keys

To override B33S's auto-generated keys, you may pass secret and access keys explicitly as environment variables. B33S server also allows regular strings as access and secret keys.

#### GNU/Linux and macOS (custom access and secret keys)

```sh
docker run \
  -p 9000:9000 \
  -p 9001:9001 \
  --name minio1 \
  -e "MINIO_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "MINIO_ROOT_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
  -v /mnt/data:/data \
  quay.io/minio/minio server /data --console-address ":9001"
```

#### Windows (custom access and secret keys)

```powershell
docker run \
  -p 9000:9000 \
  -p 9001:9001 \
  --name minio1 \
  -e "MINIO_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "MINIO_ROOT_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
  -v D:\data:/data \
  quay.io/minio/minio server /data --console-address ":9001"
```

### Run B33S Docker as a regular user

Docker provides standardized mechanisms to run docker containers as non-root users.

#### GNU/Linux and macOS (regular user)

On Linux and macOS you can use `--user` to run the container as regular user.

> NOTE: make sure --user has write permission to *${HOME}/data* prior to using `--user`.

```sh
mkdir -p ${HOME}/data
docker run \
  -p 9000:9000 \
  -p 9001:9001 \
  --user $(id -u):$(id -g) \
  --name minio1 \
  -e "MINIO_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "MINIO_ROOT_PASSWORD=wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY" \
  -v ${HOME}/data:/data \
  quay.io/minio/minio server /data --console-address ":9001"
```

#### Windows (regular user)

On windows you would need to use [Docker integrated windows authentication](https://success.docker.com/article/modernizing-traditional-dot-net-applications#integratedwindowsauthentication) and [Create a container with Active Directory Support](https://blogs.msdn.microsoft.com/containerstuff/2017/01/30/create-a-container-with-active-directory-support/)

> NOTE: make sure your AD/Windows user has write permissions to *D:\data* prior to using `credentialspec=`.

```powershell
docker run \
  -p 9000:9000 \
  -p 9001:9001 \
  --name minio1 \
  --security-opt "credentialspec=file://myuser.json"
  -e "MINIO_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "MINIO_ROOT_PASSWORD=wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY" \
  -v D:\data:/data \
  quay.io/minio/minio server /data --console-address ":9001"
```

### B33S Custom Access and Secret Keys using Docker secrets

To override B33S's auto-generated keys, you may pass secret and access keys explicitly by creating access and secret keys as [Docker secrets](https://docs.docker.com/engine/swarm/secrets/). B33S server also allows regular strings as access and secret keys.

```
echo "AKIAIOSFODNN7EXAMPLE" | docker secret create access_key -
echo "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" | docker secret create secret_key -
```

Create a B33S service using `docker service` to read from Docker secrets.

```
docker service create --name="minio-service" --secret="access_key" --secret="secret_key" quay.io/minio/minio server /data
```

Read more about `docker service` [here](https://docs.docker.com/engine/swarm/how-swarm-mode-works/services/)

#### B33S Custom Access and Secret Key files

To use other secret names follow the instructions above and replace `access_key` and `secret_key` with your custom names (e.g. `my_secret_key`,`my_custom_key`). Run your service with

```
docker service create --name="minio-service" \
  --secret="my_access_key" \
  --secret="my_secret_key" \
  --env="MINIO_ROOT_USER_FILE=my_access_key" \
  --env="MINIO_ROOT_PASSWORD_FILE=my_secret_key" \
  quay.io/minio/minio server /data
```

`MINIO_ROOT_USER_FILE` and `MINIO_ROOT_PASSWORD_FILE` also support custom absolute paths, in case Docker secrets are mounted to custom locations or other tools are used to mount secrets into the container. For example, HashiCorp Vault injects secrets to `/vault/secrets`. With the custom names above, set the environment variables to

```
MINIO_ROOT_USER_FILE=/vault/secrets/my_access_key
MINIO_ROOT_PASSWORD_FILE=/vault/secrets/my_secret_key
```

### Retrieving Container ID

To use Docker commands on a specific container, you need to know the `Container ID` for that container. To get the `Container ID`, run

```sh
docker ps -a
```

`-a` flag makes sure you get all the containers (Created, Running, Exited). Then identify the `Container ID` from the output.

### Starting and Stopping Containers

To start a stopped container, you can use the [`docker start`](https://docs.docker.com/engine/reference/commandline/start/) command.

```sh
docker start <container_id>
```

To stop a running container, you can use the [`docker stop`](https://docs.docker.com/engine/reference/commandline/stop/) command.

```sh
docker stop <container_id>
```

### B33S container logs

To access B33S logs, you can use the [`docker logs`](https://docs.docker.com/engine/reference/commandline/logs/) command.

```sh
docker logs <container_id>
```

### Monitor B33S Docker Container

To monitor the resources used by B33S container, you can use the [`docker stats`](https://docs.docker.com/engine/reference/commandline/stats/) command.

```sh
docker stats <container_id>
```

## Explore Further

* [Distributed B33S Quickstart Guide](https://min.io/docs/minio/container/operations/install-deploy-manage/deploy-minio-single-node-multi-drive.html)
* [B33S Erasure Code QuickStart Guide](https://min.io/docs/minio/container/operations/concepts/erasure-coding.html)
