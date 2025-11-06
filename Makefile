.PHONY: install
install:
	@uv sync

.PHONY: check
check:
	@echo "Checking lock file..."
	@uv lock --locked
	@echo "Linting..."
	@uv run ruff check
	@echo "Format checking..."
	@uv run ruff format --check
	@echo "Static type checking..."
	@uv run mypy

.PHONY: test
test:
	@uv run python -m pytest --cov --cov-config=pyproject.toml --cov-report=xml

.PHONY: run
run:
	@uv run python src/accountant_bot/main.py
