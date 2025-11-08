output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = aws_ecr_repository.bot.repository_url
}

output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.bot.arn
}

output "webhook_url" {
  description = "URL of the API Gateway webhook endpoint"
  value       = "${aws_api_gateway_stage.prod.invoke_url}/webhook"
}
