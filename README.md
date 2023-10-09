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

## Working With Secrets

You may override values in a `Secret` if the keys match with those in the module. For example the module `networking/infoblox` includes a secret with two fields `INFOBLOX_USERNAME` & `INFOBLOX_PASSWORD`. You can set your own values, effectively overriding them with the following:

```yaml
kind: Banana
apiVersion: konf.io/v1alpha1
modules:
- name: networking/infoblox
  secrets:
  - INFOBLOX_USERNAME=admin
  - INFOBLOX_PASSWORD=myownpassword
```

However you may not want to store the flattened (built) manifests in Git for obvious reasons. `banana` has built-in support for `sops`. By providing the `--age` command line flag, banana will encrypt the secrets so that they can be stored securely. For example

```bash
# Encrypt the bundle
banana build --age age1geawfzgrvdv5v8kd28wq8a34vvqg3zcztx76h9du95d5m62s0qhsgkrqlg > bundle-secure.yaml
# Decrypt the bundle with sops
sops --decrypt bundle-secure.yaml
```

## Getting startet

Download banana from [Releases](https://github.com/middlewaregruppen/banana/releases)