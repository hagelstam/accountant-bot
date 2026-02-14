# ECR repository
resource "aws_ecr_repository" "bot" {
  name                 = local.name_prefix
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = local.common_tags
}

# ECR lifecycle policy to keep only recent images
resource "aws_ecr_lifecycle_policy" "bot" {
  repository = aws_ecr_repository.bot.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 5 images"
        selection = {
          tagStatus     = "tagged"
          tagPrefixList = ["v"]
          countType     = "imageCountMoreThan"
          countNumber   = 5
        }
        action = {
          type = "expire"
        }
      },
      {
        rulePriority = 2
        description  = "Remove untagged images after 7 days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 7
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# SQS FIFO queue for expenses
resource "aws_sqs_queue" "expenses" {
  name                        = "${local.name_prefix}-expenses.fifo"
  fifo_queue                  = true
  content_based_deduplication = true
  visibility_timeout_seconds  = 180
  message_retention_seconds   = 86400

  tags = local.common_tags
}

# CloudWatch log group for worker Lambda
resource "aws_cloudwatch_log_group" "worker" {
  name              = "/aws/lambda/${local.name_prefix}-worker"
  retention_in_days = var.log_retention_days
  tags              = local.common_tags
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
          "${aws_cloudwatch_log_group.worker.arn}",
          "${aws_cloudwatch_log_group.worker.arn}:*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = aws_sqs_queue.expenses.arn
      }
    ]
  })
}

# Worker Lambda function
resource "aws_lambda_function" "worker" {
  function_name = "${local.name_prefix}-worker"
  role          = aws_iam_role.lambda.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.bot.repository_url}:${var.image_tag}"
  architectures = ["arm64"]

  memory_size = 128
  timeout     = 30

  environment {
    variables = {
      TELEGRAM_BOT_TOKEN      = var.telegram_bot_token
      GOOGLE_CREDENTIALS_JSON = var.google_credentials_json
      GOOGLE_SPREADSHEET_ID   = var.google_spreadsheet_id
      LOG_LEVEL               = "INFO"
    }
  }

  depends_on = [
    aws_cloudwatch_log_group.worker,
    aws_iam_role_policy.lambda
  ]

  tags = local.common_tags
}

# SQS event source mapping for worker Lambda
resource "aws_lambda_event_source_mapping" "sqs_worker" {
  event_source_arn        = aws_sqs_queue.expenses.arn
  function_name           = aws_lambda_function.worker.arn
  batch_size              = 1
  function_response_types = ["ReportBatchItemFailures"]
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

# IAM role for API Gateway to send messages to SQS
resource "aws_iam_role" "api_gateway_sqs" {
  name = "${local.name_prefix}-api-gateway-sqs-role"

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

resource "aws_iam_role_policy" "api_gateway_sqs" {
  name = "${local.name_prefix}-api-gateway-sqs-policy"
  role = aws_iam_role.api_gateway_sqs.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "sqs:SendMessage"
        Resource = aws_sqs_queue.expenses.arn
      }
    ]
  })
}

# API Gateway integration with SQS
resource "aws_api_gateway_integration" "sqs" {
  rest_api_id             = aws_api_gateway_rest_api.bot.id
  resource_id             = aws_api_gateway_resource.webhook.id
  http_method             = aws_api_gateway_method.webhook_post.http_method
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = "arn:aws:apigateway:${data.aws_region.current.id}:sqs:path/${data.aws_caller_identity.current.account_id}/${local.name_prefix}-expenses.fifo"
  credentials             = aws_iam_role.api_gateway_sqs.arn

  request_parameters = {
    "integration.request.header.Content-Type" = "'application/x-www-form-urlencoded'"
  }

  request_templates = {
    "application/json" = "Action=SendMessage&MessageGroupId=expenses&MessageBody=$util.urlEncode($input.body)"
  }
}

# API Gateway method response
resource "aws_api_gateway_method_response" "webhook_200" {
  rest_api_id = aws_api_gateway_rest_api.bot.id
  resource_id = aws_api_gateway_resource.webhook.id
  http_method = aws_api_gateway_method.webhook_post.http_method
  status_code = "200"
}

# API Gateway integration response
resource "aws_api_gateway_integration_response" "webhook_200" {
  rest_api_id = aws_api_gateway_rest_api.bot.id
  resource_id = aws_api_gateway_resource.webhook.id
  http_method = aws_api_gateway_method.webhook_post.http_method
  status_code = aws_api_gateway_method_response.webhook_200.status_code

  response_templates = {
    "application/json" = "{\"status\": \"ok\"}"
  }

  depends_on = [aws_api_gateway_integration.sqs]
}

# API Gateway deployment
resource "aws_api_gateway_deployment" "bot" {
  rest_api_id = aws_api_gateway_rest_api.bot.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.webhook.id,
      aws_api_gateway_method.webhook_post.id,
      aws_api_gateway_integration.sqs.id,
      aws_api_gateway_integration_response.webhook_200.id,
      aws_api_gateway_rest_api_policy.bot.policy,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [
    aws_api_gateway_integration.sqs
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

