# Scripts

Automation helpers for CI/CD pipelines, local bootstrap, and operational tooling will live here.

## Available Scripts

- `deploy_minikube.sh` – Bootstraps (or reuses) a `minikube` profile named
  `qubit-bots`, ensures the `trading` namespace exists, and applies the manifests
  in [`infra/k8s`](../infra/k8s) so the platform services can be exercised
  inside a local Kubernetes cluster. Customize CPU, memory, driver, or namespace
  values through environment variables described in the script header.
