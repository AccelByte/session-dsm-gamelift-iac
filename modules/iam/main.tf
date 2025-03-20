resource "aws_iam_user" "session_dsm" {
  name = "session-dsm"
  path = "/"
}

resource "aws_iam_user_policy" "gamelift_manage_game_sessions" {
  name = "gamelift-manage-game-sessions"
  user = aws_iam_user.session_dsm.name

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Action" : [
          "gamelift:ListAliases",
        ],
        "Resource" : [
          "*",
        ]
      },
      {
        "Effect" : "Allow",
        "Action" : [
          "gamelift:TerminateGameSession",
        ],
        "Resource" : [
          "*",
        ]
      },
      {
        "Effect" : "Allow",
        "Action" : [
          "gamelift:CreateGameSession",
        ],
        "Resource" : [
          "arn:aws:gamelift:*:${var.aws_account_id}:gamesession/*"
        ],
      },
      {
        "Effect" : "Allow",
        "Action" : [
          "gamelift:StartGameSessionPlacement",
          "gamelift:StopGameSessionPlacement",
        ],
        "Resource" : [
          "arn:aws:gamelift:*:${var.aws_account_id}:gamesessionqueue/*"
        ],
      },
    ]
  })
}

resource "aws_iam_access_key" "session_dsm" {
  user = aws_iam_user.session_dsm.name
}
