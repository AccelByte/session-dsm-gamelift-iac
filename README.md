# Infrastructure-as-Code for Supporting Amazon GameLift Servers Queues

Amazon GameLift Servers offers a robust queue system, which can be used to place game sessions on fleets dynamically based on location, latency, and cost. When using queues, AccelByte Gaming Services (AGS) must be informed when a queue placement completes. Amazon GameLift Servers emits session placement events automatically through Amazon EventBridge, which can be handled through an AWS Lambda to inform AccelByte about the status of the session placement.

This repo provides the setup needed to integrate AccelByte and Amazon GameLift Servers. It defines EventBridge rules to place matching events onto an SQS queue, and defines a Lambda to process events that have been placed on the SQS queue.

## Lambda

The Lambda code is provided in this repo under the `lambda` directory. The Lambda reads events from the SQS queue and, for each event, uses the AccelByte SDK to inform AGS of the status of the session placement using UpdateDSInformation.

The Lambda is capable of handling both successful and failed session placements, though it is only used for failed placements in this integration. Successful placements are expected to call UpdateDSInformation from the dedicated server in response to getting the OnStartGameSession notification from Amazon GameLift Servers.

> [!NOTE]
> For testing purposes, the Lambda assumes that the queue name exactly matches the AccelByte namespace. To support multiple queues, you may wish to design a queue naming convention that includes the AccelByte namespace and modify the Lambda to support your naming convension

> [!NOTE]  
> You must build the Lambda with the provided `build.sh` script before running `terraform apply`. Failing to do so will cause the `terraform apply` command to fail.

## Modules

### IAM

The `iam` module defines a set of credentials that are preconfigured to have the permissions needed by the Session DSM to manage Amazon GameLift Servers game sessions. These permissions include listing aliases, creating game sessions, terminating game sessions, starting session placements, and stopping session placements.

These credentials can be retrieved by running `terraform output gamelift_aws_access_key` and `terraform output gamelift_aws_secret_key` after `terraform apply`. For more information about how to configure the Session DSM using these credentials, refer to the QUICKSTART guide.

### EventBridge

The `eventbridge` module handles the creation of an AWS EventBridge rule that is responsible for filtering and routing certain events to SQS for processing. 

Amazon GameLift Servers will automatically post events to EventBridge by default, and including queue placement events. By default, it will send events about pending, failed, and succeeded queue placements. For this integration, we only want the Lambda to process failed queue placements, we apply a filter to only process events of type `PlacementFailed`, `PlacementTimedOut`, and `PlacementCancelled`. A list of game session placement events is available in the [Amazon GameLift Servers Hosting Guide](https://docs.aws.amazon.com/gamelift/latest/developerguide/queue-events.html).

This module also defines a routing rule to place matching events onto the SQS queue, where they can be processed by the Lambda later. A list of all supported EventBridge targets can be found in the [Amazon EventBridge User Guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-targets.html).


### SQS

The `sqs` module handles creation of the SQS queue that processes EventBridge events. The queue is configured as a regular, non-FIFO queue. This module also defines all the necessary permissions for the queue that will allow the Lambda to pull from it.

### Lambda

The `lambda` module defines the creation of the Lambda resource in AWS. Many of the definitions in this module are setting up AWS permissions between the SQS queue and the Lambda. The function and source mapping are defined in the resources named `lambda` and `sqs_trigger`. Note that the `filename` field in `lambda` matches the output of the `./lambda/build.sh`. By providing `source_code_hash`, these definitions guarantee that updates to the Lambda will trigger a redeploy, and similarly, running `terraform apply` without code changes will not trigger a full redeploy.

### SSM

The Lambda needs AccelByte credentials so that it can use the AccelByte SDK to make the UpdateDSInformation call. The preferred way to configure this is by adding Amazon Systems Manager Parameter Store. In the Terraform repo, three variables are configured: `/lambda/ab_base_url`, `/lambda/ab_client_id`, and `/lambda/ab_client_secret`.

Once you have applied the Terraform configuration to your AWS account, you can navigate to the [AWS SSM Parameter Store](https://us-west-2.console.aws.amazon.com/systems-manager/parameters?region=us-west-2&tab=Table#), and fill in the parameters. You will need to create AccelByte credentials with Session READ and UPDATE permissions. If using `Shared Cloud` tier, the IAM client will need `Session` → `Game Session` → read and update permissions. If using `Private Cloud` tier, the IAM client will need the `NAMESPACE:{namespace}:SESSION:GAME [READ]` and `ADMIN:NAMESPACE:{namespace}:SESSION:GAME [UPDATE]` permissions.

## Additional Resources

The Lambda uses the AccelByte Golang SDK, which can be found [here](https://github.com/AccelByte/accelbyte-go-sdk). Specifically, it makes the call to `UpdateDSInformation`, which can be found in the AccelByte API Explorer [here](https://docs.accelbyte.io/api-explorer/#Session/adminUpdateDSInformation). An example is provided in the AccelByte Golang SDK [here](https://github.com/AccelByte/accelbyte-go-sdk/blob/2abb6fb0bd663b85b687bc8122bb1aab5aa7940e/samples/cli/cmd/session/gameSession/adminUpdateDSInformation.go).

For more information about Amazon GameLift Servers Queues, refer to the official Developer Guides:

- [Managing game session placement with Amazon GameLift Servers queues](https://docs.aws.amazon.com/gamelift/latest/developerguide/queues-intro.html)
- [Set up event notification for game session placement](https://docs.aws.amazon.com/gamelift/latest/developerguide/queue-notification.html)
