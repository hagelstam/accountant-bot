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
  default     = "latest"
}
