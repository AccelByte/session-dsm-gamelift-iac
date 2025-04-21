package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/session"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth"
	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclient/game_session"
	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclientmodels"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const (
	// GameLift constants
	PlacementCancelled = "PlacementCancelled"
	PlacementTimedOut  = "PlacementTimedOut"
	PlacementFailed    = "PlacementFailed"

	// AccelByte constants
	DsSourceAws             = "AWS"
	DsStatusAvailable       = "AVAILABLE"
	DsStatusFailedToRequest = "FAILED_TO_REQUEST"

	// Terraform-defined constants
	SsmParamNameBaseUrl       = "/lambda/ab_base_url"
	SsmParamNameClientId      = "/lambda/ab_client_id"
	SsmParamNameClientSecret  = "/lambda/ab_client_secret"
	SsmParamNameNamespaceName = "/lambda/ab_namespace_name"
)

// See https://docs.aws.amazon.com/gamelift/latest/developerguide/queue-events.html
type GameLiftEventDetail struct {
	Type        string `json:"type"` // PlacementFulfilled, PlacementCancelled, PlacementTimedOut, PlacementFailed
	PlacementId string `json:"placementId"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`

	// NOTE: the following fields are only provided when `Type` is `PlacementFulfilled`
	GameSessionArn       string `json:"gameSessionArn,omitempty"`
	GameSessionRegion    string `json:"gameSessionRegion,omitempty"`
	DnsName              string `json:"dnsName,omitempty"`
	IpAddress            string `json:"ipAddress,omitempty"`
	Port                 string `json:"port,omitempty"`
	PlacedPlayerSessions []any  `json:"placedPlayerSessions,omitempty"`
}

func getAccelByteCredentialsFromSsm(ctx context.Context, ssmClient *ssm.Client) (repository.ConfigRepository, error) {
	accelByteBaseUrl, err := getParamFromSsm(ctx, ssmClient, SsmParamNameBaseUrl)
	if err != nil {
		return nil, err
	}

	accelByteClientId, err := getParamFromSsm(ctx, ssmClient, SsmParamNameClientId)
	if err != nil {
		return nil, err
	}

	accelByteClientSecret, err := getParamFromSsm(ctx, ssmClient, SsmParamNameClientSecret)
	if err != nil {
		return nil, err
	}

	return &auth.ConfigRepositoryImpl{
		BaseUrl:      accelByteBaseUrl,
		ClientId:     accelByteClientId,
		ClientSecret: accelByteClientSecret,
	}, nil
}

func getParamFromSsm(ctx context.Context, ssmClient *ssm.Client, paramName string) (string, error) {
	withDecryption := true
	paramValue, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &paramName,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return "", err
	}

	if paramValue.Parameter == nil {
		return "", fmt.Errorf("nil parameter returned from SSM")
	}

	if paramValue.Parameter.Value == nil {
		return "", fmt.Errorf("nil parameter value returned from SSM")
	}

	return *paramValue.Parameter.Value, nil
}

func getAccelByteGameSessionService(configRepo repository.ConfigRepository) (*session.GameSessionService, error) {
	oauthService := iam.OAuth20Service{
		Client:                 factory.NewIamClient(configRepo),
		ConfigRepository:       configRepo,
		TokenRepository:        auth.DefaultTokenRepositoryImpl(),
		RefreshTokenRepository: auth.DefaultRefreshTokenImpl(),
	}

	clientId := configRepo.GetClientId()
	clientSecret := configRepo.GetClientSecret()
	err := oauthService.LoginClient(&clientId, &clientSecret)
	if err != nil {
		return nil, err
	}

	return &session.GameSessionService{
		Client:           factory.NewSessionClient(oauthService.ConfigRepository),
		ConfigRepository: oauthService.ConfigRepository,
		TokenRepository:  oauthService.TokenRepository,
	}, nil
}

func main() {
	lambda.Start(func(ctx context.Context, sqsEvent events.SQSEvent) error {
		awsConfig, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			log.Printf("Error loading AWS config: %v", err)
			return err
		}
		awsSsmClient := ssm.NewFromConfig(awsConfig)

		accelByteCredentials, err := getAccelByteCredentialsFromSsm(ctx, awsSsmClient)
		if err != nil {
			log.Printf("Error getting Accelbyte credentials from SSM: %v", err)
			return err
		}

		accelByteNamespace, err := getParamFromSsm(ctx, awsSsmClient, SsmParamNameNamespaceName)
		if err != nil {
			return err
		}

		accelByteSessionService, err := getAccelByteGameSessionService(accelByteCredentials)
		if err != nil {
			log.Printf("Error creating AccelByte session client: %v", err)
			return err
		}

		defaultAccelByteRetryPolicy := &utils.Retry{
			MaxTries:   1,
			Backoff:    utils.NewConstantBackoff(0),
			Transport:  accelByteSessionService.Client.Runtime.Transport,
			RetryCodes: utils.RetryCodes,
		}

		for _, record := range sqsEvent.Records {
			var eventBridgeEvent events.EventBridgeEvent
			err := json.Unmarshal([]byte(record.Body), &eventBridgeEvent)
			if err != nil {
				log.Printf("Skipping event with invalid JSON: %v", err)
				continue
			}

			// We are passed the GameLift Queue ARN from which the event was spawned in the 'Resouces' array
			if len(eventBridgeEvent.Resources) == 0 {
				log.Printf("Skipping event with no resources: %v", eventBridgeEvent)
				continue
			}

			var detail GameLiftEventDetail
			err = json.Unmarshal(eventBridgeEvent.Detail, &detail)
			if err != nil {
				log.Printf("Skipping event with invalid detail JSON: %v", eventBridgeEvent.Detail)
				continue
			}

			getSessionInput := &game_session.GetGameSessionParams{
				Context:     ctx,
				Namespace:   accelByteNamespace,
				SessionID:   detail.PlacementId,
				RetryPolicy: defaultAccelByteRetryPolicy,
			}
			sessionRes, err := accelByteSessionService.GetGameSessionShort(getSessionInput)
			if err != nil {
				log.Printf("Error getting Accelbyte session information: %v", err)
				continue
			}

			// NOTE: UpdateDSInformation requires connection data, even for failed requests
			source := DsSourceAws
			ip := "127.0.0.1"
			port := int32(7777)
			deployment := detail.Type
			region := eventBridgeEvent.Region

			var status string
			if detail.Type == PlacementCancelled || detail.Type == PlacementFailed || detail.Type == PlacementTimedOut {
				status = DsStatusFailedToRequest
			} else {
				status = DsStatusAvailable
				ip = detail.IpAddress

				parsedPort, err := strconv.ParseInt(detail.Port, 10, 32)
				if err != nil {
					fmt.Printf("failed to parse port: %s. Defaulting to %d.", err, port)
				} else {
					port = int32(parsedPort)
				}

				deployment = detail.GameSessionArn
				region = detail.GameSessionRegion
			}

			updateDsInfoInput := &game_session.AdminUpdateDSInformationParams{
				Context:     ctx,
				Namespace:   accelByteNamespace,
				SessionID:   detail.PlacementId,
				RetryPolicy: defaultAccelByteRetryPolicy,
				Body: &sessionclientmodels.ApimodelsUpdateGamesessionDSInformationRequest{
					ClientVersion: sessionRes.Configuration.ClientVersion,
					Region:        &eventBridgeEvent.Region,
					CreatedRegion: &region,
					Deployment:    &deployment,
					Source:        &source,
					ServerID:      &detail.PlacementId,
					Status:        &status,
					GameMode:      sessionRes.MatchPool,
					IP:            &ip,
					Port:          &port,
				},
			}

			err = accelByteSessionService.AdminUpdateDSInformationShort(updateDsInfoInput)
			if err != nil {
				log.Printf("Error updating session info for failed session placement %s: %v", detail.PlacementId, err)
				continue
			}

			log.Printf("Successfully updated DS information through AccelByte. Status: %s", status)
		}

		return nil
	})
}
