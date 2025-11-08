import json
from unittest.mock import MagicMock, Mock, patch

import pytest

from accountant_bot.expense_parser import Expense
from accountant_bot.sheets_service import SheetsService


@pytest.fixture
def mock_credentials() -> str:
    return json.dumps({
        "type": "service_account",
        "project_id": "test-project",
        "private_key_id": "test-key-id",
        "private_key": "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----\n",
        "client_email": "test@test-project.iam.gserviceaccount.com",
        "client_id": "123456789",
        "auth_uri": "https://accounts.google.com/o/oauth2/auth",
        "token_uri": "https://oauth2.googleapis.com/token",
    })


class TestSheetsService:
    @patch("accountant_bot.sheets_service.Credentials.from_service_account_info")
    @patch("accountant_bot.sheets_service.gspread.authorize")
    def test_create_client_success(
        self, mock_authorize: Mock, mock_from_service_account: Mock, mock_credentials: str
    ) -> None:
        mock_creds = MagicMock()
        mock_from_service_account.return_value = mock_creds

        mock_client = MagicMock()
        mock_authorize.return_value = mock_client

        service = SheetsService(credentials_json=mock_credentials, spreadsheet_id="test-sheet-id")

        assert service.spreadsheet_id == "test-sheet-id"
        assert service.client == mock_client
        mock_authorize.assert_called_once_with(mock_creds)

    def test_create_client_invalid_json(self) -> None:
        with pytest.raises(ValueError):
            SheetsService(credentials_json="invalid json", spreadsheet_id="test-sheet-id")

    @patch("accountant_bot.sheets_service.Credentials.from_service_account_info")
    @patch("accountant_bot.sheets_service.gspread.authorize")
    def test_get_current_month_worksheet(
        self, mock_authorize: Mock, mock_from_service_account: Mock, mock_credentials: str
    ) -> None:
        mock_from_service_account.return_value = MagicMock()

        mock_client = MagicMock()
        mock_authorize.return_value = mock_client

        mock_worksheet = MagicMock()
        mock_spreadsheet = MagicMock()
        mock_spreadsheet.worksheets.return_value = [mock_worksheet]
        mock_client.open_by_key.return_value = mock_spreadsheet

        service = SheetsService(credentials_json=mock_credentials, spreadsheet_id="test-sheet-id")
        worksheet = service._get_current_month_worksheet()

        assert worksheet == mock_worksheet
        mock_client.open_by_key.assert_called_with("test-sheet-id")

    @patch("accountant_bot.sheets_service.Credentials.from_service_account_info")
    @patch("accountant_bot.sheets_service.gspread.authorize")
    def test_get_current_month_worksheet_no_worksheets(
        self, mock_authorize: Mock, mock_from_service_account: Mock, mock_credentials: str
    ) -> None:
        mock_from_service_account.return_value = MagicMock()

        mock_client = MagicMock()
        mock_authorize.return_value = mock_client

        mock_spreadsheet = MagicMock()
        mock_spreadsheet.worksheets.return_value = []
        mock_client.open_by_key.return_value = mock_spreadsheet

        service = SheetsService(credentials_json=mock_credentials, spreadsheet_id="test-sheet-id")

        with pytest.raises(ValueError):
            service._get_current_month_worksheet()

    @patch("accountant_bot.sheets_service.Credentials.from_service_account_info")
    @patch("accountant_bot.sheets_service.gspread.authorize")
    def test_find_next_empty_row(
        self, mock_authorize: Mock, mock_from_service_account: Mock, mock_credentials: str
    ) -> None:
        mock_from_service_account.return_value = MagicMock()
        mock_client = MagicMock()
        mock_authorize.return_value = mock_client

        service = SheetsService(credentials_json=mock_credentials, spreadsheet_id="test-sheet-id")

        mock_worksheet = MagicMock()
        mock_worksheet.col_values.return_value = [
            "Jobb",
            "Studiestöd",
            "Total Net income ",
            "Fundamentals",
            "Turun energia",
            "",
        ]

        next_row = service._find_next_empty_row(mock_worksheet)
        assert next_row == 6

    @patch("accountant_bot.sheets_service.Credentials.from_service_account_info")
    @patch("accountant_bot.sheets_service.gspread.authorize")
    def test_add_expense(self, mock_authorize: Mock, mock_from_service_account: Mock, mock_credentials: str) -> None:
        mock_from_service_account.return_value = MagicMock()

        mock_client = MagicMock()
        mock_authorize.return_value = mock_client

        mock_worksheet = MagicMock()
        mock_worksheet.col_values.return_value = [
            "Jobb",
            "Total Net income ",
            "Fundamentals",
            "",
        ]

        mock_spreadsheet = MagicMock()
        mock_spreadsheet.worksheets.return_value = [mock_worksheet]
        mock_client.open_by_key.return_value = mock_spreadsheet

        service = SheetsService(credentials_json=mock_credentials, spreadsheet_id="test-sheet-id")

        expense = Expense(desc="Coffee", amount=3.50)
        service.add_expense(expense)

        mock_worksheet.update_cell.assert_any_call(4, 1, "Coffee")
        mock_worksheet.update_cell.assert_any_call(4, 2, 3.50)
