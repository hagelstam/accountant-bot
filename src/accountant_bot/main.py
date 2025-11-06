import logging

from telegram.ext import ApplicationBuilder

from accountant_bot.config import TELEGRAM_BOT_TOKEN, get_logging_level
from accountant_bot.handlers import register_handlers

logging.basicConfig(level=get_logging_level())


def main() -> None:
    app = ApplicationBuilder().token(TELEGRAM_BOT_TOKEN).build()
    register_handlers(app)
    print("🤖 Accountant bot is running...")
    app.run_polling()


if __name__ == "__main__":
    main()
