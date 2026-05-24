# Architecture

BYO Addons is a small Kubernetes operator project. It defines an `AddonSet` custom resource and reconciles each enabled addon component into a desired-state `ConfigMap`.

## Core Flow

1. A user applies an `AddonSet`.
2. Kubernetes stores it through the `platform.byoaddons.io/v1alpha1` CRD.
3. The controller watches the `AddonSet`.
4. The controller adds a finalizer.
5. For each enabled component, the controller creates or updates an owned desired-state `ConfigMap`.
6. The controller updates `AddonSet.status`.
7. When the `AddonSet` is deleted, the controller removes owned ConfigMaps and then removes the finalizer.

## Tooling

- Kubebuilder-style layout: API, controller, manager, CRD, RBAC, and Kustomize config.
- Kustomize: local/raw deployment path for CRD, RBAC, and manager manifests.
- Helm: packaged and configurable operator installation.
- Argo CD: optional GitOps example for installing the operator and addon charts.

The current version records addon intent. It does not directly install addon Helm charts yet.
