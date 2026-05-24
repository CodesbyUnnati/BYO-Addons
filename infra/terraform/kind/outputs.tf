output "argocd_namespace" {
  value = kubernetes_namespace.argocd.metadata[0].name
}

output "byo_addons_namespace" {
  value = kubernetes_namespace.byo_addons.metadata[0].name
}
