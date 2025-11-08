"""Script to register Telegram webhook."""

import argparse
import asyncio
import logging
import sys

from telegram import Bot

from accountant_bot.config import TELEGRAM_BOT_TOKEN, TELEGRAM_SECRET_TOKEN

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)


async def set_webhook(webhook_url: str, secret_token: str | None = None) -> None:
    """
    Set the Telegram webhook URL.

    Args:
        webhook_url: The HTTPS URL for the webhook
        secret_token: Optional secret token for additional security
    """
    bot = Bot(token=TELEGRAM_BOT_TOKEN)

    try:
        # Delete existing webhook first
        await bot.delete_webhook(drop_pending_updates=True)
        logger.info("Deleted existing webhook")

        # Set new webhook
        success = await bot.set_webhook(
            url=webhook_url,
            secret_token=secret_token,
            allowed_updates=["message", "callback_query", "inline_query"],
            drop_pending_updates=True,
        )

        if success:
            logger.info(f"✅ Webhook successfully set to: {webhook_url}")

            # Verify webhook info
            webhook_info = await bot.get_webhook_info()
            logger.info(f"Webhook info: {webhook_info}")
        else:
            logger.error("❌ Failed to set webhook")
            sys.exit(1)

    except Exception:
        logger.exception("Error setting webhook")
        sys.exit(1)


async def delete_webhook() -> None:
    """Delete the current webhook."""
    bot = Bot(token=TELEGRAM_BOT_TOKEN)

    try:
        await bot.delete_webhook(drop_pending_updates=True)
        logger.info("✅ Webhook successfully deleted")
    except Exception:
        logger.exception("Error deleting webhook")
        sys.exit(1)


async def get_webhook_info() -> None:
    """Get current webhook information."""
    bot = Bot(token=TELEGRAM_BOT_TOKEN)

    try:
        webhook_info = await bot.get_webhook_info()
        logger.info(f"Current webhook info:\n{webhook_info}")
    except Exception:
        logger.exception("Error getting webhook info")
        sys.exit(1)


def main() -> None:
    """Main entry point for webhook setup."""
    parser = argparse.ArgumentParser(description="Manage Telegram webhook")
    parser.add_argument(
        "action",
        choices=["set", "delete", "info"],
        help="Action to perform (set, delete, or info)",
    )
    parser.add_argument(
        "--url",
        help="Webhook URL (required for 'set' action)",
    )

    args = parser.parse_args()

    if args.action == "set":
        if not args.url:
            logger.error("--url is required for 'set' action")
            sys.exit(1)

        asyncio.run(set_webhook(args.url, TELEGRAM_SECRET_TOKEN))

    elif args.action == "delete":
        asyncio.run(delete_webhook())

    elif args.action == "info":
        asyncio.run(get_webhook_info())


if __name__ == "__main__":
    main()
