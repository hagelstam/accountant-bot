variable "environment" {
  type        = string
  description = "Environment name"
  default     = "prod"
}

variable "telegram_bot_token" {
  type        = string
  description = "Telegram bot token"
  sensitive   = true
}

variable "google_credentials_json" {
  type        = string
  description = "Google credentials JSON"
  sensitive   = true
}

variable "google_spreadsheet_id" {
  type        = string
  description = "Google spreadsheet ID"
  sensitive   = true
}

variable "image_tag" {
  type        = string
  description = "Docker image tag to deploy"
}

variable "log_retention_days" {
  type        = number
  description = "CloudWatch log retention in days"
  default     = 30
}

variable "tags" {
  type        = map(string)
  description = "Tags to apply to all resources"
  default     = {}
}
