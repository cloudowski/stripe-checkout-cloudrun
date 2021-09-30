variable "container_image" {
  type = string
}
variable "project" {
  type = string
}

variable "region" {
  type = string
}

variable "service_name" {
  type = string
}
variable "dns_domain" {
  type    = string
  default = ""
}

variable "config" {
  type = list(map(string))
}
