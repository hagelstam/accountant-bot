"""Telegram bot message handlers."""

import logging

from telegram import Update
from telegram.ext import Application, CommandHandler, ContextTypes

logger = logging.getLogger(__name__)


async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle the /start command"""
    if update.effective_user and update.message:
        user = update.effective_user
        logger.info(f"User {user.id} ({user.username}) started the bot")

        await update.message.reply_text(f"Hi {user.first_name}! 👋\n\n")


async def echo(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Echo back any text message"""
    if update.message and update.message.text:
        logger.info(f"Received message: {update.message.text}")
        await update.message.reply_text(f"You said: {update.message.text}")


def register_handlers(application: Application) -> None:
    application.add_handler(CommandHandler("start", start))

    from telegram.ext import MessageHandler, filters

    application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, echo))

    logger.info("Handlers registered successfully")
