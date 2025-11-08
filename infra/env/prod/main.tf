variable "telegram_bot_token" {
  type        = string
  description = "Telegram bot token"
  sensitive   = true
}

variable "image_tag" {
  type        = string
  description = "Docker image tag to deploy"
  default     = "latest"
}

module "bot" {
  source = "../../modules/bot"

  environment        = "prod"
  image_tag          = var.image_tag
  telegram_bot_token = var.telegram_bot_token
  log_retention_days = 7

  tags = {
    Project     = "accountant-bot"
    Environment = "prod"
  }
}
