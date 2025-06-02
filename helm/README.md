# Venowolf minlog

## Preface

Starting with Loki v2.8, TSDB is the recommended Loki index, and, object storage has better scalability than filesystem. Both time series DB and object storage are IO-intensive applications. High-performance storage solutions are a prerequisite for good log management.

Compared with network storage, local volumes have better performance. K8s supports hostPath and local persistent storage (https://kubernetes.io/docs/concepts/storage/volumes/#local).

This directory contains documentation for deploying minlog with helm. It is split into the following parts:

* `local-static-volume/`: deploy persistent volume(https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner.git).
Deploy local static storage and support dynamic PVC based on block devices.
* `local-storage/`: deploy persistent volume(k8s local volumes)
Must set a PersistentVolume nodeAffinity when using local volumes.
* `hostpath/`: deploy persistent volume(k8s hostpath)
Create local path for loki
* `loki-sc.yaml`: 

## Preview the website

Run `make docs`.
This launches a preview of the website with the current grafana docs at `http://localhost:3002/docs/alloy/latest/` which automatically refreshes when changes are made to content in the `sources` directory.
Make sure Docker is running.

## Update CloudWatch docs

First, inside the `docs/` folder run `make check-cloudwatch-integration` to verify that the CloudWatch docs needs updating.

If the check fails, then the doc supported services list should be updated.
For that, run `make generate-cloudwatch-integration` to get the updated list, which should replace the old one in [the docs](./sources/static/configuration/integrations/cloudwatch-exporter-config.md).

## Update generated reference docs

Some sections of Grafana Alloy reference documentation are automatically generated. To update them, run `make generate-docs`.