# kmaint

[![Release](https://github.com/amimof/kmaint/actions/workflows/release.yaml/badge.svg)](https://github.com/amimof/kmaint/actions/workflows/release.yaml)

`kmaint` is a command line utility that generates Kubernetes configuration from a declarative specification

**Work in progress** *kmaint is still under active development and most features are still in an idéa phase. Please check in from time to time for follow the progress* 🧡


## The `kmaint.yaml` file

This file describes how you application will ultimately look like. Kmaint will generate either `Kustomize` or `Helm` manifests (depending on the module) and place everything in `src/`. 

```yaml
kind: Konf
apiVersion: konf.io/v1alpha1
name: integration
modules:
- name: monitoring/grafana
  opts:
    version: 1.4.4
    namespace: infra-monitoring
- name: ingress/traefik
  opts:
    version: 2.2.2
- name: auth/dex
  opts:
    version: 4.1.1
    namespace: kube-system
```

## Getting startet

Download kmaint from [Releases](https://github.com/amimof/kmaint/releases)