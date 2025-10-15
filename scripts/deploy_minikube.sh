#!/usr/bin/env bash
set -euo pipefail

MINIKUBE_PROFILE=${MINIKUBE_PROFILE:-qubit-bots}
MINIKUBE_DRIVER=${MINIKUBE_DRIVER:-docker}
MINIKUBE_CPUS=${MINIKUBE_CPUS:-4}
MINIKUBE_MEMORY=${MINIKUBE_MEMORY:-8192}
K8S_NAMESPACE=${K8S_NAMESPACE:-trading}
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
MANIFEST_DIR="${ROOT_DIR}/infra/k8s"

command -v minikube >/dev/null 2>&1 || {
  echo "minikube is required but was not found in PATH" >&2
  exit 1
}

command -v kubectl >/dev/null 2>&1 || {
  echo "kubectl is required but was not found in PATH" >&2
  exit 1
}

echo "Checking minikube profile '${MINIKUBE_PROFILE}'..."
if ! minikube status -p "${MINIKUBE_PROFILE}" >/dev/null 2>&1; then
  echo "Starting minikube (driver=${MINIKUBE_DRIVER}, cpus=${MINIKUBE_CPUS}, memory=${MINIKUBE_MEMORY}MB)..."
  minikube start \
    -p "${MINIKUBE_PROFILE}" \
    --driver="${MINIKUBE_DRIVER}" \
    --cpus="${MINIKUBE_CPUS}" \
    --memory="${MINIKUBE_MEMORY}"
else
  echo "minikube profile '${MINIKUBE_PROFILE}' already running."
fi

export KUBECONFIG="$(minikube kubeconfig -p "${MINIKUBE_PROFILE}")"

kubectl config use-context "${MINIKUBE_PROFILE}" >/dev/null 2>&1 || true

if ! kubectl get namespace "${K8S_NAMESPACE}" >/dev/null 2>&1; then
  echo "Creating namespace '${K8S_NAMESPACE}'..."
  kubectl create namespace "${K8S_NAMESPACE}"
else
  echo "Namespace '${K8S_NAMESPACE}' already exists."
fi

echo "Looking for Kubernetes manifests in ${MANIFEST_DIR}..."
mapfile -t manifest_files < <(find "${MANIFEST_DIR}" -maxdepth 1 -type f \( -name '*.yml' -o -name '*.yaml' \))

if ((${#manifest_files[@]} == 0)); then
  echo "No manifest files found. Skipping apply step."
else
  apply_manifests=false
  for file in "${manifest_files[@]}"; do
    if grep -Eq '^[[:space:]]*kind:' "${file}"; then
      apply_manifests=true
      break
    fi
  done

  if [[ "${apply_manifests}" == true ]]; then
    echo "Applying Kubernetes manifests from ${MANIFEST_DIR}..."
    kubectl apply -n "${K8S_NAMESPACE}" -f "${MANIFEST_DIR}"
    echo "Deployment initiated. Current resource status:"
    kubectl get all -n "${K8S_NAMESPACE}"
  else
    echo "Manifest files appear to be placeholders (no 'kind:' found). Skipping apply step."
  fi
fi

echo "To interact with the services, use: minikube service --namespace ${K8S_NAMESPACE} <service-name>"
