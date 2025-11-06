import logging
from typing import Any

from telegram import Update
from telegram.ext import CommandHandler, ContextTypes, MessageHandler, filters

logger = logging.getLogger(__name__)


async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle /start command."""
    if not update.message:
        return None

    await update.message.reply_text(
        "*Hello👋*\n\nWelcome to the Accountant bot.",
        parse_mode="Markdown",
    )


async def echo(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Echo any message back to the user."""
    message = update.message

    if not message or not message.text:
        return None

    await message.reply_text(f"You said: {message.text}")


def register_handlers(app: Any) -> None:
    """Register handlers."""
    app.add_handler(CommandHandler("start", start))
    app.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, echo))
