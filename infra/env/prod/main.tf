module "bot" {
  source = "../../modules/bot"

  environment             = "prod"
  image_tag               = var.image_tag
  telegram_bot_token      = var.telegram_bot_token
  google_credentials_json = var.google_credentials_json
  google_spreadsheet_id   = var.google_spreadsheet_id
  log_retention_days      = 3

  tags = {
    Project     = "accountant-bot"
    Environment = "prod"
  }
}
