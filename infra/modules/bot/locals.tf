locals {
  project_name = "accountant-bot"
  name_prefix  = "${local.project_name}-${var.environment}"
  common_tags  = var.tags
  source_path  = "${path.module}/../../../src/accountant_bot"

  # Telegram IP ranges
  # https://core.telegram.org/bots/webhooks#the-short-version
  telegram_ip_ranges = [
    "149.154.160.0/20",
    "91.108.4.0/22"
  ]
}
