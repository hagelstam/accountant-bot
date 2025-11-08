import logging

from pydantic_settings import BaseSettings


class Config(BaseSettings):
    telegram_bot_token: str
    logging_level: str = "INFO"


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
