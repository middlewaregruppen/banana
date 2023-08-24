# banana

[![Release](https://github.com/middlewaregruppen/banana/actions/workflows/release.yaml/badge.svg)](https://github.com/middlewaregruppen/banana/actions/workflows/release.yaml)

`banana` is a command line utility that generates Kubernetes configuration from a declarative specification

**Work in progress** *banana is still under active development and most features are still in an idéa phase. Please check in from time to time for follow the progress* 🧡


## The `banana.yaml` file

This file describes how your application will ultimately look like. For example

```yaml
kind: Banana
apiVersion: konf.io/v1alpha1
modules:
- name: monitoring/grafana
  components:
  - dashboards
  - loki
- name: ingress/nginx
  components:
  - tls
- name: auth/dex
  version: v3.1.14
```

Then build with `banana build`

## Getting startet

Download banana from [Releases](https://github.com/middlewaregruppen/banana/releases)