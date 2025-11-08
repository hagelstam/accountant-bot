"""AWS Lambda handler for Telegram webhook."""

import json
import logging
from typing import Any

from telegram import Update
from telegram.ext import ApplicationBuilder

from accountant_bot.config import TELEGRAM_BOT_TOKEN, get_logging_level
from accountant_bot.handlers import register_handlers

logging.basicConfig(
    level=get_logging_level(),
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

# Initialize the bot application once (Lambda container reuse)
application = ApplicationBuilder().token(TELEGRAM_BOT_TOKEN).build()
register_handlers(application)


async def process_update(event: dict[str, Any]) -> dict[str, Any]:
    """Process incoming Telegram update."""
    try:
        body = event.get("body", "")
        update_data = json.loads(body) if isinstance(body, str) else body

        update = Update.de_json(update_data, application.bot)
        if update:
            await application.initialize()
            await application.process_update(update)
            logger.info(f"Successfully processed update {update.update_id}")
        else:
            logger.warning("Received invalid update")

        return {
            "statusCode": 200,
            "body": json.dumps({"status": "ok"}),
        }

    except json.JSONDecodeError:
        logger.exception("JSON decode error")
        return {
            "statusCode": 400,
            "body": json.dumps({"error": "Invalid JSON"}),
        }
    except Exception:
        logger.exception("Error processing update")
        return {
            "statusCode": 500,
            "body": json.dumps({"error": "Internal server error"}),
        }


def lambda_handler(event: dict[str, Any], context: Any) -> dict[str, Any]:
    """
    AWS Lambda handler for Telegram webhook.

    Args:
        event: API Gateway event containing the webhook request
        context: Lambda context object

    Returns:
        API Gateway response with status code and body
    """
    import asyncio

    logger.info(f"Received event: {json.dumps(event)}")

    # Run the async handler
    loop = asyncio.get_event_loop()
    return loop.run_until_complete(process_update(event))
