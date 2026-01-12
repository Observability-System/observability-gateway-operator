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

### Prerequisites
- Docker with BuildKit support
- docker `buildx` enabled
- Access to a container registry

### Build the Image Locally
Build a single-architecture image and tag it locally:

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

## Development
1. Build & run locally (great for development):
```bash
make install   # Install CRD locally
make run       # Run operator on your machine
```

2. Build & push image
```bash
IMG=alexandrosst/observability-gateway-operator:v0.1.1 make docker-build docker-push
```

3. Test full installation from your repo
```bash
kubectl apply -k github.com/alexandrosst/observability-gateway-operator/config/default
```

## Uninstallation
```bash
kubectl delete -k https://github.com/alexandrosst/observability-gateway-operator/config/default?ref=v0.1.1
kubectl delete crd observabilitygateways.observability.x-k8s.io
```

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/observability-gateway-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/observability-gateway-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/observability-gateway-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/observability-gateway-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v2-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

