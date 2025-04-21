resource "aws_sqs_queue" "queue" {
  name                       = var.queue_name
  message_retention_seconds  = var.message_retention_seconds
  visibility_timeout_seconds = var.visibility_timeout_seconds
}

resource "aws_lambda_permission" "sqs_lambda" {
  statement_id  = "AllowSQSInvokeLambda"
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_arn
  principal     = "sqs.amazonaws.com"
  source_arn    = aws_sqs_queue.queue.arn
}

resource "aws_sqs_queue_policy" "sqs_policy" {
  queue_url = aws_sqs_queue.queue.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        },
        Action    = "SQS:SendMessage"
        Resource  = aws_sqs_queue.queue.arn
      },
      {
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Action   = "SQS:ReceiveMessage"
        Resource = aws_sqs_queue.queue.arn
      }
    ]
  })
}
