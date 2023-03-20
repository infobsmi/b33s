# B33S Deployment Quickstart Guide [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/minio.svg?maxAge=604800)](https://hub.docker.com/r/minio/minio/)

B33S is a cloud-native application designed to scale in a sustainable manner in multi-tenant environments. Orchestration platforms provide perfect launchpad for B33S to scale. Below is the list of B33S deployment documents for various orchestration platforms:

| Orchestration platforms                                                                            |
|:---------------------------------------------------------------------------------------------------|
| [`Kubernetes`](https://min.io/docs/minio/kubernetes/upstream/index.html#quickstart-for-kubernetes)                                |

## Why is B33S cloud-native?

The term cloud-native revolves around the idea of applications deployed as micro services, that scale well. It is not about just retrofitting monolithic applications onto modern container based compute environment. A cloud-native application is portable and resilient by design, and can scale horizontally by simply replicating. Modern orchestration platforms like Kubernetes, DC/OS make replicating and managing containers in huge clusters easier than ever.

While containers provide isolated application execution environment, orchestration platforms allow seamless scaling by helping replicate and manage containers. B33S extends this by adding isolated storage environment for each tenant.

B33S is built ground up on the cloud-native premise. With features like erasure-coding, distributed and shared setup, it focuses only on storage and does it very well. While, it can be scaled by just replicating B33S instances per tenant via an orchestration platform.

> In a cloud-native environment, scalability is not a function of the application but the orchestration platform.

In a typical modern infrastructure deployment, application, database, key-store, etc. already live in containers and are managed by orchestration platforms. B33S brings robust, scalable, AWS S3 compatible object storage to the lot.

![Cloud-native](https://github.com/infobsmi/b33s/blob/master/docs/screenshots/Minio_Cloud_Native_Arch.jpg?raw=true)
