resource "aws_cloudwatch_event_rule" "gamelift_placement_events" {
  name           = var.rule_name
  description    = "Rule to capture GameLift Queue Placement Events for failed placements"
  event_bus_name = var.event_bus_name

  event_pattern = jsonencode({
    "source" : ["aws.gamelift"],
    "detail-type" : ["GameLift Queue Placement Event"],
    "detail" : {
      "type" : [
        "PlacementFailed",
        "PlacementTimedOut",
        "PlacementCancelled",
      ]
    }
  })
}

resource "aws_cloudwatch_event_target" "send_to_sqs" {
  rule = aws_cloudwatch_event_rule.gamelift_placement_events.name
  arn  = var.sqs_queue_arn
}
