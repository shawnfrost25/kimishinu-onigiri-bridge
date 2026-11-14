variable "region" {
  type        = string
  description = "The region where the instances are going to be located"
}

variable "author" {
  type        = string
  description = "The author of the AWS credentials"
}

variable "backend_instance_type" {
  type        = string
  description = "The backend instance type"
}
