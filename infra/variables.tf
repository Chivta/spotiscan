variable "resource_group_name" {
  description = "Name of the Azure resource group"
  type        = string
  default     = "spotiscan-rg"
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "polandcentral"
}

variable "cluster_name" {
  description = "AKS cluster name"
  type        = string
  default     = "spotiscan-aks"
}

variable "node_count" {
  description = "Number of nodes in the default pool"
  type        = number
  default     = 2
}

variable "node_vm_size" {
  description = "VM size for AKS nodes"
  type        = string
  default     = "standard_b2s_v2"
}

variable "subscription_id" {
  description = "Azure subscription ID"
  type        = string
}
