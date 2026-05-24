variable "kubeconfig_path" {
  description = "Path to the kubeconfig for the target cluster."
  type        = string
  default     = "~/.kube/config"
}

variable "argocd_namespace" {
  description = "Namespace where Argo CD should be installed."
  type        = string
  default     = "argocd"
}

variable "install_argocd" {
  description = "Install Argo CD with the Helm provider."
  type        = bool
  default     = true
}
