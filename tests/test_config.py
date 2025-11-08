import logging
from unittest.mock import patch

from accountant_bot.config import get_logging_level


def test_get_logging_level_debug():
    with patch("accountant_bot.config.LOGGING_LEVEL", "DEBUG"):
        assert get_logging_level() == logging.DEBUG
