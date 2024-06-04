package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

const defaultRegion = "us-east-1"

var brc *bedrockruntime.Client

func init() {

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = defaultRegion
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	brc = bedrockruntime.NewFromConfig(cfg)
}

const modelID = "anthropic.claude-3-sonnet-20240229-v1:0"

func main() {

	reader := bufio.NewReader(os.Stdin)

	converseInput := &bedrockruntime.ConverseInput{
		ModelId: aws.String(modelID),
	}

	for {
		fmt.Print("\nEnter your message: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		userMsg := types.Message{
			Role: types.ConversationRoleUser,
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{
					Value: input,
				},
			},
		}

		converseInput.Messages = append(converseInput.Messages, userMsg)
		output, err := brc.Converse(context.Background(), converseInput)

		if err != nil {
			log.Fatal(err)
		}

		reponse, _ := output.Output.(*types.ConverseOutputMemberMessage)
		responseContentBlock := reponse.Value.Content[0]
		text, _ := responseContentBlock.(*types.ContentBlockMemberText)

		fmt.Println(text.Value)

		assistantMsg := types.Message{
			Role:    types.ConversationRoleAssistant,
			Content: reponse.Value.Content,
		}

		converseInput.Messages = append(converseInput.Messages, assistantMsg)
	}
}
