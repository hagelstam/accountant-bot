import re
from dataclasses import dataclass


@dataclass
class Expense:
    desc: str
    amount: float


def parse_expense(message: str) -> Expense | None:
    if not message or not message.strip():
        return None

    pattern = r"^(.+?)\s+([\d,.]+)$"
    match = re.match(pattern, message.strip())

    if not match:
        return None

    desc = match.group(1).strip()
    amount_str = match.group(2).replace(",", ".")

    try:
        amount = float(amount_str)
        if amount <= 0:
            return None
        return Expense(desc=desc, amount=amount)
    except ValueError:
        return None
