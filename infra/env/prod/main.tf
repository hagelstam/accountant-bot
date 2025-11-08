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

module "bot" {
  source = "../../modules/bot"

  environment             = "prod"
  image_tag               = var.image_tag
  telegram_bot_token      = var.telegram_bot_token
  google_credentials_json = var.google_credentials_json
  google_spreadsheet_id   = var.google_spreadsheet_id
  log_retention_days      = 7

  tags = {
    Project     = "accountant-bot"
    Environment = "prod"
  }
}
