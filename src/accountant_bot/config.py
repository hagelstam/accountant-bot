import logging

from pydantic_settings import BaseSettings, SettingsConfigDict


class Config(BaseSettings):
    telegram_bot_token: str
    logging_level: str = "INFO"

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")


config = Config()  # type: ignore[call-arg]

TELEGRAM_BOT_TOKEN = config.telegram_bot_token
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
