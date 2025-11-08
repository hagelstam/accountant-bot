# Accountant Bot

[![Production](https://img.shields.io/github/actions/workflow/status/hagelstam/accountant-bot/ci-cd.yml?branch=main)](https://github.com/hagelstam/accountant-bot/actions/workflows/ci-cd.yml?query=branch%3Amain)
[![codecov](https://codecov.io/gh/hagelstam/accountant-bot/branch/main/graph/badge.svg)](https://codecov.io/gh/hagelstam/accountant-bot)
[![License](https://img.shields.io/github/license/hagelstam/accountant-bot)](https://img.shields.io/github/license/hagelstam/accountant-bot)

A Telegram bot for managing personal finances, deployed on AWS Lambda with API Gateway using webhook-based architecture.

## 🏗️ Architecture

- **Compute**: AWS Lambda (containerized Python 3.12)
- **API**: API Gateway HTTP API with webhook endpoint
- **Security**: AWS WAF with Telegram IP whitelist + secret token validation
- **Container Registry**: Amazon ECR with image scanning and lifecycle policies
- **Infrastructure**: Terraform for IaC
- **CI/CD**: GitHub Actions with separate build and deploy workflows

## 🚀 Quick Start

### Local Development

```bash
# 1. Copy environment template and fill in your bot token
cp env.template .env

# 2. Install dependencies (for testing/linting)
make install

# 3. Run tests and checks
make check
make test

# 4. Run bot locally with docker-compose (recommended)
make run

# Or run natively (requires exporting env vars first)
export $(cat .env | xargs) && make run-native
```

### Production Deployment

The bot is automatically deployed to AWS Lambda when changes are pushed to the `main` branch.

## 📋 Prerequisites

### For Local Development

- Docker and Docker Compose
- Python 3.14+ (optional, only needed for native execution)
- [uv](https://github.com/astral-sh/uv) package manager (optional, for testing/linting)
- Telegram bot token (get from [@BotFather](https://t.me/botfather))

### For AWS Deployment

- AWS Account with appropriate permissions
- AWS CLI configured
- Terraform 1.13.5+
- Docker for building images
- GitHub repository secrets configured (see below)

## 🔧 Setup

### 1. Create Telegram Bot

1. Talk to [@BotFather](https://t.me/botfather) on Telegram
2. Use `/newbot` command
3. Save the bot token

### 2. Configure Local Environment

```bash
# Copy the environment template
cp env.template .env

# Edit .env and add your bot token
# Get token from @BotFather
```

**Important**: The `.env` file is gitignored and should never be committed. It's only used locally. In production, environment variables are managed by AWS through Terraform.

### 3. Generate Secret Token

```bash
# Generate a secure random token for webhook validation
openssl rand -hex 32
```

Add this to your `.env` file for local development and to GitHub secrets for production.

### 4. Configure GitHub Secrets

Add these secrets to your GitHub repository (Settings → Secrets → Actions):

- `TELEGRAM_BOT_TOKEN`: Your bot token from BotFather
- `TELEGRAM_SECRET_TOKEN`: The generated secret token

### 5. Configure AWS Access

The GitHub Actions workflows use OIDC to authenticate with AWS. Ensure your AWS account has:

- An IAM role named `deployer` with the ARN: `arn:aws:iam::747683189254:role/deployer`
- Trust policy allowing GitHub Actions OIDC
- Permissions for ECR, Lambda, API Gateway, CloudWatch, WAF, and S3

### 6. Create S3 Backend for Terraform

```bash
# Create S3 bucket for Terraform state
aws s3 mb s3://tfstate-accountant-bot --region eu-north-1
aws s3api put-bucket-versioning \
  --bucket tfstate-accountant-bot \
  --versioning-configuration Status=Enabled
```

## 🚢 Deployment

### Automated Deployment (Recommended)

Push to `main` branch to trigger the full CI/CD pipeline:

```bash
git push origin main
```

This will:

1. ✅ Run tests and code quality checks
2. 🐳 Build and push Docker image to ECR
3. 🏗️ Deploy infrastructure with Terraform
4. 🔗 Configure Telegram webhook

### Manual Deployment

#### Full Deployment

```bash
# Set environment variables
export TELEGRAM_BOT_TOKEN="your-token"
export TELEGRAM_SECRET_TOKEN="your-secret"

# Deploy everything (build + push + infrastructure + webhook)
make deploy
```

#### Quick Deployment (Infrastructure Only)

```bash
# Deploy infrastructure changes only (uses existing Docker image)
make deploy-quick
```

#### Custom Image Tag

```bash
# Deploy specific image version
make deploy IMAGE_TAG=v20240101-abc1234
```

## 🐳 Docker Commands

```bash
# Build Docker image
make docker-build

# Build and push to ECR
make docker-push

# Build with custom tag
make docker-build IMAGE_TAG=v1.0.0
```

## 🏗️ Infrastructure Commands

```bash
# Initialize Terraform
make tf-init

# Plan infrastructure changes
make tf-plan

# Apply infrastructure changes
make tf-apply

# View outputs (webhook URL, etc.)
make tf-output

# Format Terraform files
make tf-fmt

# Destroy infrastructure (⚠️ use with caution)
make tf-destroy
```

## 🪝 Webhook Management

```bash
# Set webhook URL (automatically done during deployment)
make webhook-set

# Get webhook information
make webhook-info

# Delete webhook (switches to polling mode)
make webhook-delete
```

## 📊 Monitoring

```bash
# Follow Lambda logs in real-time
make logs

# View recent logs (last hour)
make logs-recent

# Test Lambda function directly
make lambda-invoke-test
```

## 🧪 Testing

```bash
# Run all tests with coverage
make test

# Run code quality checks
make check

# Run specific test file
uv run pytest tests/test_config.py -v
```

## 🔒 Security Features

### 1. IP Whitelisting

- AWS WAF blocks all requests except from Telegram IP ranges
- IP ranges: `149.154.160.0/20`, `91.108.4.0/22`

### 2. Secret Token Validation

- All webhook requests must include `X-Telegram-Bot-Api-Secret-Token` header
- Token is validated in Lambda before processing

### 3. Rate Limiting

- WAF enforces 2000 requests per 5 minutes per IP

### 4. Container Security

- ECR image scanning on push
- Trivy vulnerability scanning in CI/CD
- Regular base image updates

## 📁 Project Structure

```
.
├── src/
│   └── accountant_bot/
│       ├── __init__.py
│       ├── config.py              # Configuration management
│       ├── handlers.py            # Telegram message handlers
│       └── lambda_handler.py      # AWS Lambda webhook entry point
├── scripts/
│   └── setup_webhook.py           # Webhook configuration script
├── infra/
│   ├── modules/
│   │   └── bot/                   # Reusable Terraform module
│   │       ├── main.tf
│   │       ├── variables.tf
│   │       └── outputs.tf
│   └── env/
│       └── prod/                  # Production environment
│           ├── main.tf
│           └── providers.tf
├── .github/
│   └── workflows/
│       ├── ci-cd.yml              # Main CI/CD pipeline
│       ├── build-and-push.yml     # Docker image build/push
│       ├── deploy.yml             # Terraform deployment
│       ├── validate-py.yml        # Python validation
│       └── validate-tf.yml        # Terraform validation
├── tests/                         # Test files
├── Dockerfile                     # Lambda-compatible container
├── Makefile                       # Development and deployment commands
└── pyproject.toml                 # Python dependencies and config
```

## 🛠️ Development Workflow

1. **Create feature branch**

   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make changes and test locally**

   ```bash
   # Ensure .env file is configured
   make check
   make test
   make run  # Runs in docker-compose
   ```

3. **Create pull request**

   - Automated checks will run (Python validation, Terraform validation)
   - Review Terraform plan in PR comments

4. **Merge to main**
   - Full CI/CD pipeline deploys to production
   - Bot is automatically updated with zero downtime

## 📝 Environment Variables

### Local Development (via .env file)

```bash
TELEGRAM_BOT_TOKEN=your_token        # Required: Bot token from BotFather
TELEGRAM_SECRET_TOKEN=your_secret    # Optional: For webhook validation
ENVIRONMENT=development              # Default: development
LOGGING_LEVEL=INFO                   # Default: INFO
```

**Note**: Create a `.env` file from `env.template`. This file is gitignored and used only for local development with docker-compose.

### Production (AWS Lambda - set by Terraform)

- `TELEGRAM_BOT_TOKEN`: Bot token (from GitHub secrets)
- `TELEGRAM_SECRET_TOKEN`: Webhook secret (from GitHub secrets)
- `WEBHOOK_URL`: API Gateway webhook URL (auto-generated by Terraform)
- `ENVIRONMENT`: Runtime environment (`production`)
- `LOGGING_LEVEL`: Log level (`INFO`)

All production environment variables are injected by Terraform into the Lambda function. No `.env` files are used in production.

## 🐛 Troubleshooting

### Bot not responding

1. Check webhook status:

   ```bash
   make webhook-info
   ```

2. View recent logs:

   ```bash
   make logs-recent
   ```

3. Test Lambda directly:
   ```bash
   make lambda-invoke-test
   ```

### Deployment fails

1. Check Terraform state:

   ```bash
   make tf-output
   ```

2. Verify AWS credentials:

   ```bash
   aws sts get-caller-identity
   ```

3. Check GitHub Actions logs in repository

### Local development issues

1. Verify `.env` file exists and is configured:

   ```bash
   cat .env
   # Should show your TELEGRAM_BOT_TOKEN and other variables
   ```

2. Rebuild docker images:

   ```bash
   make dev-down
   docker-compose build --no-cache
   make run
   ```

3. For native execution, ensure dependencies are installed:

   ```bash
   make clean
   make install
   ```

## 📚 Additional Resources

- [Telegram Bot API Documentation](https://core.telegram.org/bots/api)
- [AWS Lambda Documentation](https://docs.aws.amazon.com/lambda/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [python-telegram-bot Documentation](https://docs.python-telegram-bot.org/)

## 📄 License

[MIT License](LICENSE)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Open a Pull Request
