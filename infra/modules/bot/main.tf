# Generate requirements.txt from uv.lock
resource "null_resource" "requirements" {
  triggers = {
    uv_lock = filemd5("${path.module}/../../../uv.lock")
  }

  provisioner "local-exec" {
    working_dir = "${path.module}/../../.."
    command     = "uv export --frozen --no-hashes --no-dev > requirements.txt"
  }
}

# IAM role for Lambda
resource "aws_iam_role" "lambda" {
  name = "${local.name_prefix}-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = local.common_tags
}

# IAM policy for Lambda
resource "aws_iam_role_policy" "lambda" {
  name = "${local.name_prefix}-lambda-policy"
  role = aws_iam_role.lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = [
          "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${local.name_prefix}",
          "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${local.name_prefix}:*"
        ]
      }
    ]
  })
}

# Lambda function
module "bot" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "~> 7.21.1"

  depends_on = [null_resource.requirements]

  function_name = local.name_prefix
  handler       = "accountant_bot.lambda_handler.lambda_handler"
  runtime       = "python3.14"
  architectures = ["arm64"]

  memory_size = 512
  timeout     = 30

  source_path = [
    {
      path             = local.source_path
      pip_requirements = "${path.module}/../../../requirements.txt"
    }
  ]

  environment_variables = var.lambda_environment_variables

  create_lambda_function_url = false
  create_role                = false
  lambda_role                = aws_iam_role.lambda.arn

  cloudwatch_logs_retention_in_days = var.log_retention_days

  tags = local.common_tags
}

# CloudWatch log group for API Gateway
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${local.name_prefix}"
  retention_in_days = var.log_retention_days
  tags              = local.common_tags
}

# IAM role for API Gateway logging
resource "aws_iam_role" "api_gateway_cloudwatch" {
  name = "${local.name_prefix}-api-gateway-cloudwatch-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "apigateway.amazonaws.com"
        }
      }
    ]
  })

  tags = local.common_tags
}

# Attach policy for API Gateway logging
resource "aws_iam_role_policy_attachment" "api_gateway_cloudwatch" {
  role       = aws_iam_role.api_gateway_cloudwatch.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonAPIGatewayPushToCloudWatchLogs"
}

# API Gateway account settings for logging
resource "aws_api_gateway_account" "main" {
  cloudwatch_role_arn = aws_iam_role.api_gateway_cloudwatch.arn
}

# API Gateway REST API
resource "aws_api_gateway_rest_api" "bot" {
  name        = local.name_prefix
  description = "Telegram webhook API for ${local.project_name}"

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = local.common_tags
}

# API Gateway resource policy to allow only Telegram IPs
resource "aws_api_gateway_rest_api_policy" "bot" {
  rest_api_id = aws_api_gateway_rest_api.bot.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = "*"
        Action    = "execute-api:Invoke"
        Resource  = "${aws_api_gateway_rest_api.bot.execution_arn}/*"
        Condition = {
          IpAddress = {
            "aws:SourceIp" = local.telegram_ip_ranges
          }
        }
      }
    ]
  })
}

# API Gateway resource for /webhook
resource "aws_api_gateway_resource" "webhook" {
  rest_api_id = aws_api_gateway_rest_api.bot.id
  parent_id   = aws_api_gateway_rest_api.bot.root_resource_id
  path_part   = "webhook"
}

# API Gateway method (POST /webhook)
resource "aws_api_gateway_method" "webhook_post" {
  rest_api_id   = aws_api_gateway_rest_api.bot.id
  resource_id   = aws_api_gateway_resource.webhook.id
  http_method   = "POST"
  authorization = "NONE"
}

# API Gateway integration with Lambda
resource "aws_api_gateway_integration" "lambda" {
  rest_api_id             = aws_api_gateway_rest_api.bot.id
  resource_id             = aws_api_gateway_resource.webhook.id
  http_method             = aws_api_gateway_method.webhook_post.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = module.bot.lambda_function_invoke_arn
}

# API Gateway deployment
resource "aws_api_gateway_deployment" "bot" {
  rest_api_id = aws_api_gateway_rest_api.bot.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.webhook.id,
      aws_api_gateway_method.webhook_post.id,
      aws_api_gateway_integration.lambda.id,
      aws_api_gateway_rest_api_policy.bot.policy,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [
    aws_api_gateway_integration.lambda
  ]
}

# API Gateway stage
resource "aws_api_gateway_stage" "prod" {
  deployment_id = aws_api_gateway_deployment.bot.id
  rest_api_id   = aws_api_gateway_rest_api.bot.id
  stage_name    = "prod"

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      resourcePath   = "$context.resourcePath"
      status         = "$context.status"
      protocol       = "$context.protocol"
      responseLength = "$context.responseLength"
    })
  }

  depends_on = [aws_api_gateway_account.main]

  tags = local.common_tags
}

# Lambda permission for API Gateway
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = module.bot.lambda_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.bot.execution_arn}/*/*"
}
