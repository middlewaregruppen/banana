# banana

[![Release](https://github.com/middlewaregruppen/banana/actions/workflows/release.yaml/badge.svg)](https://github.com/middlewaregruppen/banana/actions/workflows/release.yaml)

`banana` is a command line utility that generates Kubernetes configuration from a declarative specification

**Work in progress** *banana is still under active development and most features are still in an idéa phase. Please check in from time to time for follow the progress* 🧡


## The `banana.yaml` file

This file describes how you application will ultimately look like. Banana will generate either `Kustomize` or `Helm` manifests (depending on the module) and place everything in `src/`. 

```yaml
kind: Banana
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

Download banana from [Releases](https://github.com/middlewaregruppen/banana/releases)