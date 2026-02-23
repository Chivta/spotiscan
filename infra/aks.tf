resource "azurerm_kubernetes_cluster" "main" {
  name = var.cluster_name

  location = azurerm_resource_group.main.location

  resource_group_name = azurerm_resource_group.main.name

  dns_prefix = var.cluster_name

  default_node_pool {
    name = "default"

    node_count = var.node_count
    vm_size    = var.node_vm_size
  }

  identity {
    type = "SystemAssigned"
  }
}
