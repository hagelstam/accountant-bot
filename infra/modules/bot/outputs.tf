output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = module.bot.lambda_function_arn
}

output "webhook_url" {
  description = "URL of the API Gateway webhook endpoint"
  value       = "${aws_api_gateway_stage.prod.invoke_url}/webhook"
}
