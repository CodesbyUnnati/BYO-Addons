# Architecture

BYO Addons is a small open source platform layer for Kubernetes addon selection. It is inspired by the "bring your own CNI" idea: the cluster team owns a baseline platform, while application and platform teams can choose supported free and open source providers from a curated catalog.

## Control Plane

The operator owns the `AddonSet` custom resource:

- `spec.clusterName` identifies the target cluster.
- `spec.gitOpsEngine` names the external applier, usually Argo CD.
- `spec.components` lists desired addons such as Cilium, OpenEBS, or kube-prometheus-stack.

In version `0.1.0`, reconciliation records desired component state as owned ConfigMaps and reports status on the `AddonSet`. That is intentionally modest: it demonstrates the Kubernetes reconciliation loop without hiding too much logic. A future branch can promote this into direct Argo CD `Application` generation or Helm SDK-based installation.

## Reconciliation Flow

1. A user or Argo CD applies an `AddonSet`.
2. The controller watches the resource.
3. For each enabled component, it creates or updates a desired-state ConfigMap.
4. The controller updates component status and the aggregate `Ready` condition.
5. If the `AddonSet` is deleted, the finalizer cleans up owned desired-state ConfigMaps.

## Tooling Roles

- Kubebuilder layout: API, controller, RBAC, CRD, and manager structure.
- Helm: reusable operator installation chart.
- Kustomize: raw manifest deployment and overlays.
- Argo CD: GitOps sync engine for operator and addon manifests.
- Terraform: cluster namespace and Argo CD bootstrap.
- Ansible: repeatable post-cluster bootstrap tasks.
- Prometheus and Grafana: operator metrics, alerts, and dashboards.
