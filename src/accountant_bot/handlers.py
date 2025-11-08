import logging

from telegram import Update
from telegram.ext import Application, CommandHandler, ContextTypes

from accountant_bot.config import GOOGLE_CREDENTIALS_JSON, GOOGLE_SPREADSHEET_ID
from accountant_bot.expense_parser import parse_expense
from accountant_bot.sheets_service import SheetsService

logger = logging.getLogger(__name__)
sheets_service = SheetsService(credentials_json=GOOGLE_CREDENTIALS_JSON, spreadsheet_id=GOOGLE_SPREADSHEET_ID)


async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle the /start command"""

    if update.effective_user and update.message:
        user = update.effective_user
        logger.info(f"User {user.id} ({user.username}) started the bot")

        welcome_message = (
            f"Hi {user.first_name}! 👋\n\n"
            "I'm your personal accountant bot. Send me expenses in this format:\n\n"
            "Example: `Lunch 2.95`\n\n"
        )

        await update.message.reply_text(welcome_message)


async def handle_expense(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    """Handle expense messages and update Google Sheets"""

    if not update.message or not update.message.text:
        return

    message_text = update.message.text
    logger.info(f"Received message: {message_text}")

    expense = parse_expense(message_text)
    if not expense:
        await update.message.reply_text("Could not parse expense. Please use format:\n\nExample: `Lunch 2.95`")
        return

    try:
        sheets_service.add_expense(expense)
        await update.message.reply_text(f"Added: {expense.desc} - €{expense.amount:.2f}")
    except Exception as e:
        logger.exception("Failed to add expense")
        await update.message.reply_text(f"Failed to add expense: {e}")


def register_handlers(application: Application) -> None:
    application.add_handler(CommandHandler("start", start))

    from telegram.ext import MessageHandler, filters

    application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_expense))
    logger.info("Handlers registered successfully")
