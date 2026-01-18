module "bot" {
  source = "../../modules/bot"

  environment        = "prod"
  log_retention_days = 7

  lambda_environment_variables = {
    TELEGRAM_BOT_TOKEN      = var.telegram_bot_token
    GOOGLE_CREDENTIALS_JSON = var.google_credentials_json
    GOOGLE_SPREADSHEET_ID   = var.google_spreadsheet_id
    LOGGING_LEVEL           = "INFO"
  }

  tags = {
    Project     = "accountant-bot"
    Environment = "prod"
  }
}
