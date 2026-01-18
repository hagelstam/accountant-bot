variable "environment" {
  type        = string
  description = "Environment name"
  default     = "prod"
}

variable "lambda_environment_variables" {
  type        = map(string)
  description = "Environment variables for the Lambda function"
}

variable "log_retention_days" {
  type        = number
  description = "CloudWatch log retention in days"
  default     = 7
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply to all resources"
  default     = {}
}
