output "kube_config_command" {
  description = "Command to fetch kubeconfig"
  value       = "az aks get-credentials --resource-group ${azurerm_resource_group.main.name} --name ${azurerm_kubernetes_cluster.main.name}"
}

output "cluster_fqdn" {
  description = "AKS API server FQDN"
  value       = azurerm_kubernetes_cluster.main.fqdn
}

output "node_resource_group" {
  description = "Auto-created resource group for AKS node resources (VMs, disks, LB)"
  value       = azurerm_kubernetes_cluster.main.node_resource_group
}
