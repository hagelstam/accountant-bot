output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = aws_ecr_repository.bot.repository_url
}

output "worker_function_arn" {
  description = "ARN of the worker Lambda function"
  value       = aws_lambda_function.worker.arn
}

output "sqs_queue_url" {
  description = "URL of the SQS expenses queue"
  value       = aws_sqs_queue.expenses.url
}

output "webhook_url" {
  description = "URL of the API Gateway webhook endpoint"
  value       = "${aws_api_gateway_stage.prod.invoke_url}/webhook"
}
