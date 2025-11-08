import json
import logging

import gspread
from google.oauth2.service_account import Credentials

from accountant_bot.expense_parser import Expense

logger = logging.getLogger(__name__)


class SheetsService:
    def __init__(self, credentials_json: str, spreadsheet_id: str) -> None:
        self.spreadsheet_id = spreadsheet_id
        self.client = self._create_client(credentials_json)

    def _create_client(self, credentials_json: str) -> gspread.Client:
        try:
            creds_dict = json.loads(credentials_json)
            scopes = [
                "https://www.googleapis.com/auth/spreadsheets",
                "https://www.googleapis.com/auth/drive",
            ]
            credentials = Credentials.from_service_account_info(creds_dict, scopes=scopes)
            return gspread.authorize(credentials)
        except json.JSONDecodeError as e:
            logger.exception("Failed to parse credentials JSON")
            raise ValueError from e
        except Exception as e:
            logger.exception("Failed to create Google Sheets client")
            raise RuntimeError from e

    def _get_current_month_worksheet(self) -> gspread.Worksheet:
        spreadsheet = self.client.open_by_key(self.spreadsheet_id)
        worksheets = spreadsheet.worksheets()

        if not worksheets:
            raise ValueError()

        # The leftmost (newest) sheet is the first in the list
        return worksheets[0]

    def _find_next_empty_row(self, worksheet: gspread.Worksheet) -> int:
        # Get all values from the first column (Fundamentals)
        col_values = worksheet.col_values(1)

        # Find where expenses start (after "Total Net income")
        start_row = None
        for i, value in enumerate(col_values, start=1):
            if "Total Net income" in str(value):
                # Expenses start after the header row following income
                start_row = i + 2  # Skip the header row
                break

        if start_row is None:
            raise ValueError()

        # Find the next empty row after start_row
        for i in range(start_row, len(col_values) + 1):
            if i >= len(col_values) or not str(col_values[i - 1]).strip():
                return i

        # If all rows are filled, append to the end
        return len(col_values) + 1

    def add_expense(self, expense: Expense) -> None:
        worksheet = self._get_current_month_worksheet()
        next_row = self._find_next_empty_row(worksheet)

        worksheet.update_cell(next_row, 1, expense.desc)
        worksheet.update_cell(next_row, 2, expense.amount)
