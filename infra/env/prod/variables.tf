variable "aws_region" {
  default = "eu-north-1"
}

variable "telegram_bot_token" {
  type      = string
  sensitive = true
}

variable "google_credentials_json" {
  type      = string
  sensitive = true
}

variable "google_spreadsheet_id" {
  type      = string
  sensitive = true
}
