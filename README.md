# Observability Gateway Operator
Kubernetes Operator for managing **class-based observability gateway deployments** with declarative lifecycle management.

The operator automates the creation, update, and cleanup of OpenTelemetry gateway deployments based on a single custom resource. It is designed for environments where telemetry ingestion must be **structured**, **separated by class**, and **managed consistently over time**.

<p align='center'>
    <img src="assets/Observability-Gateway-Operator-Icon.svg" alt="Observability Gateway Operator Logo" width="300">
</p>

## Conceptual Overview
In many observability deployments, telemetry from different sources often needs to be handled differently depending on importance, volume, or service-level requirements. Common examples include:
- critical vs best-effort telemetry
- production vs development traffic
- premium vs standard tenants

The **Observability Gateway Operator** provides a **declarative Kubernetes API** that allows users to describe what *gateway classes should exist* and *how they should be provisioned*, while the operator ensures that the corresponding Kubernetes resources are continuously reconciled.

## Reconciliation Model
For each declared gateway class, the operator manages:
- One **Deployment** per priority class
- One **Service** per priority class
- Shared configuration references (ConfigMap)
- Ownership via Kubernetes ownerReferences

The reconciliation loop ensures that:
- declared classes always exist
- removed classes are garbage-collected
- replica counts are enforced
- resources are recreated after failure

All managed resources are **fully owned** by the custom resource via Kubernetes owner references, ensuring automatic garbage collection and preventing orphaned objects.

Reconciliation is triggered on creation, update, or deletion of an `ObservabilityGateway` resource, as well as when managed resources drift from the desired state.

## Installation
### Prerequisites
- kubectl version v1.11.3+
- Kubernetes v1.11.3+ cluster
- go version v1.24.6+ (for local development)

### Step 1: Install the Custom Resource Definition
The CRD must be installed before creating any `ObservabilityGateway` resources:
```bash
kubectl apply -f https://raw.githubusercontent.com/Observability-System/Observability-Gateway-Operator/main/config/crd/bases/observability.x-k8s.io_observabilitygateways.yaml
```

Verify:
```bash
kubectl get crds | grep observabilitygateways
```

### Step 2: Deploy the Operator
Deploy the operator using the default kustomize configuration:
```bash
kubectl apply -k https://github.com/Observability-System/Observability-Gateway-Operator/config/default
```
This installs:
- the controller Deployment
- required RBAC resources
- the namespace `observability-system`.

Verify:
```bash
kubectl get pods -A | grep observability-gateway
```

## Usage
### Step 1: Create a Namespace
```bash
kubectl create namespace observability
```

### Step 2: Apply Shared Configuration
Create or apply a ConfigMap containing OpenTelemetry configuration:
```bash
kubectl apply -f examples/otel-configmap.yaml
```
Ensure the name matches the one referenced in the custom resource.

### Step 3: Create a Gateway Resource
```bash
kubectl apply -f examples/gateway.yaml
```

### Step 4: Verify Managed Resources
```bash
kubectl get deployments,services -n observability
```

Expected output resembles:
```bash
prio-ingestion-gateway-gold     3/3
prio-ingestion-gateway-silver   2/2
prio-ingestion-gateway-bronze   1/1
```
Each class is managed independently.

## Build and Publish
This project provides a Makefile with helper targets for building and publishing the operator container image. To see all supported Makefile targets:

```bash
make help
```

### Build Prerequisites
- Docker with BuildKit support
- docker `buildx` enabled
- Access to a container registry

### Build the Image Locally
Build a single-architecture image:

```bash
make docker-build
```

### Push the Image
Build and push a single-architecture image:

```bash
make docker-push
```
This target builds the image and pushes it to the configured registry.

### Build and Push Multi-Architecture Image
To build and push a multi-architecture image (linux/amd64, linux/arm64):

```bash
make docker-pushx
```
This uses Docker Buildx to publish a multi-platform manifest.