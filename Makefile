PROJECT_NAME   := accountant-bot
AWS_REGION     := eu-north-1
AWS_ACCOUNT_ID := 747683189254
ECR_REPOSITORY := $(PROJECT_NAME)-prod
TERRAFORM_DIR  := infra/env/prod
IMAGE_TAG      ?= v$(shell date +%Y%m%d)-$(shell git rev-parse --short HEAD)

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
