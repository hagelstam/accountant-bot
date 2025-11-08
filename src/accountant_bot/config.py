import logging

from pydantic_settings import BaseSettings


class Config(BaseSettings):
    telegram_bot_token: str
    google_credentials_json: str
    google_spreadsheet_id: str
    logging_level: str = "INFO"


config = Config()  # type: ignore[call-arg]

TELEGRAM_BOT_TOKEN = config.telegram_bot_token
GOOGLE_CREDENTIALS_JSON = config.google_credentials_json
GOOGLE_SPREADSHEET_ID = config.google_spreadsheet_id
LOGGING_LEVEL = config.logging_level


def get_logging_level() -> int:
    levels = {
        "DEBUG": logging.DEBUG,
        "INFO": logging.INFO,
        "WARNING": logging.WARNING,
        "ERROR": logging.ERROR,
        "CRITICAL": logging.CRITICAL,
    }

    level = LOGGING_LEVEL.upper()
    if level not in levels:
        print(f"Warning: Invalid logging level '{LOGGING_LEVEL}'. Using INFO.")
        return logging.INFO

    return levels[level]
