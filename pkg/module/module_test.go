package module

var kustomizationData = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- resource.yaml`

var ingressData = `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-ingress
  namespace: test-namespace
spec:
  ingressClassName: nginx
  rules:
  - host: infra-test-ingress
    http:
      paths:
      - backend:
          service:
            name: test-service
            port:
              number: 80
        path: /
        pathType: Prefix`
