terraform {
  required_providers {
    spot = {
      source = "rackerlabs/spot"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.7.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.0.1"
    }
  }
}

variable "cloudspace_name" {
  description = "The cloudspace name"
  type        = string
}

variable "token" {
  description = "The spot token"
  type        = string
  sensitive   = true
}

provider "spot" {
  token = var.token
}

data "spot_kubeconfig" "example" {
  cloudspace_name = var.cloudspace_name
}

# Use Kubernetes provider to deploy resources
provider "kubernetes" {
  host     = data.spot_kubeconfig.example.kubeconfigs[0].host
  token    = data.spot_kubeconfig.example.kubeconfigs[0].token
  insecure = data.spot_kubeconfig.example.kubeconfigs[0].insecure
}

resource "kubernetes_namespace" "test" {
  metadata {
    name = "test"
  }
}

resource "kubernetes_deployment" "test" {
  metadata {
    name      = "test"
    namespace = kubernetes_namespace.test.metadata.0.name
  }
  spec {
    replicas = 2
    selector {
      match_labels = {
        app = "test"
      }
    }
    template {
      metadata {
        labels = {
          app = "test"
        }
      }
      spec {
        container {
          image = "hashicorp/http-echo"
          name  = "http-echo"
          args  = ["-text=test"]

          resources {
            limits = {
              memory = "512M"
              cpu    = "1"
            }
            requests = {
              memory = "256M"
              cpu    = "50m"
            }
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "test" {
  metadata {
    name      = "test-service"
    namespace = kubernetes_namespace.test.metadata.0.name
  }
  spec {
    selector = {
      app = kubernetes_deployment.test.metadata.0.name
    }

    port {
      port = 5678
    }
  }
}

# Use Helm provider to deploy resources
provider "helm" {
  kubernetes {
    host  = data.spot_kubeconfig.example.kubeconfigs[0].host
    token = data.spot_kubeconfig.example.kubeconfigs[0].token
  }
}

resource "helm_release" "nginx_ingress" {
  name      = "nginx-ingress-controller"
  namespace = kubernetes_namespace.test.metadata.0.name

  repository = "https://charts.bitnami.com/bitnami"
  chart      = "nginx-ingress-controller"

  set {
    name  = "service.type"
    value = "LoadBalancer"
  }
  set {
    name  = "service.annotations.service\\.beta\\.kubernetes\\.io/do-loadbalancer-name"
    value = format("%s-nginx-ingress", var.cloudspace_name)
  }
}

resource "kubernetes_ingress_v1" "test_ingress" {
  wait_for_load_balancer = true
  metadata {
    name      = "test-ingress"
    namespace = kubernetes_namespace.test.metadata.0.name
    annotations = {
      "kubernetes.io/ingress.class"          = "nginx"
      "ingress.kubernetes.io/rewrite-target" = "/"
    }
  }

  spec {
    rule {
      http {
        path {
          backend {
            service {
              name = kubernetes_service.test.metadata.0.name
              port {
                number = 5678
              }
            }
          }

          path = "/test"
        }
      }
    }
  }
}