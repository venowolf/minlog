# Venowolf minlog

## Preface

Starting with Loki v2.8, TSDB is the recommended Loki index, and, object storage has better scalability than filesystem. Both time series DB and object storage are IO-intensive applications. High-performance storage solutions are a prerequisite for good log management.

This directory contains documentation for deploying minlog with helm. It is split into the following parts:

* `persistentvolume/`: deploying local persistent volume.
  Local volumes are preferred for better performance. However, if the cluster already has high-performance storage (such as distributed storage or cloud storage), there is no need to deploy local volumes again.

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