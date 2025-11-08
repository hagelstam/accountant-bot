from accountant_bot.expense_parser import Expense, parse_expense


class TestParseExpense:
    def test_parse_simple_expense(self) -> None:
        result = parse_expense("Lunch 2.95")
        assert result is not None
        assert result.desc == "Lunch"
        assert result.amount == 2.95

    def test_parse_expense_with_multiple_words(self) -> None:
        result = parse_expense("Gym membership 31.99")
        assert result is not None
        assert result.desc == "Gym membership"
        assert result.amount == 31.99

    def test_parse_expense_with_comma_decimal(self) -> None:
        result = parse_expense("Groceries 15,50")
        assert result is not None
        assert result.desc == "Groceries"
        assert result.amount == 15.5

    def test_parse_expense_whole_number(self) -> None:
        result = parse_expense("Movie ticket 12")
        assert result is not None
        assert result.desc == "Movie ticket"
        assert result.amount == 12.0

    def test_parse_expense_extra_whitespace(self) -> None:
        result = parse_expense("  Lunch   2.95  ")
        assert result is not None
        assert result.desc == "Lunch"
        assert result.amount == 2.95

    def test_parse_expense_empty_string(self) -> None:
        assert parse_expense("") is None
        assert parse_expense("   ") is None

    def test_parse_expense_missing_amount(self) -> None:
        assert parse_expense("Lunch") is None

    def test_parse_expense_invalid_amount(self) -> None:
        assert parse_expense("Lunch abc") is None

    def test_parse_expense_negative_amount(self) -> None:
        assert parse_expense("Lunch -2.95") is None

    def test_parse_expense_zero_amount(self) -> None:
        assert parse_expense("Lunch 0") is None
        assert parse_expense("Lunch 0.00") is None

    def test_expense_dataclass(self) -> None:
        expense = Expense(desc="Test", amount=10.0)
        assert expense.desc == "Test"
        assert expense.amount == 10.0
