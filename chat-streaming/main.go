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

	converseStreamInput := &bedrockruntime.ConverseStreamInput{
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

		converseStreamInput.Messages = append(converseStreamInput.Messages, userMsg)

		output, err := brc.ConverseStream(context.Background(), converseStreamInput)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Print("[Assistant]: ")

		assistantMsg, err := processStreamingOutput(output, func(ctx context.Context, part string) error {
			fmt.Print(part)
			return nil
		})

		if err != nil {
			log.Fatal("streaming output processing error: ", err)
		}

		converseStreamInput.Messages = append(converseStreamInput.Messages, assistantMsg)

		fmt.Println()
	}
}

type StreamingOutputHandler func(ctx context.Context, part string) error

func processStreamingOutput(output *bedrockruntime.ConverseStreamOutput, handler StreamingOutputHandler) (types.Message, error) {

	var combinedResult string

	msg := types.Message{}

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ConverseStreamOutputMemberMessageStart:

			msg.Role = v.Value.Role

		case *types.ConverseStreamOutputMemberContentBlockDelta:

			textResponse := v.Value.Delta.(*types.ContentBlockDeltaMemberText)
			handler(context.Background(), textResponse.Value)
			combinedResult = combinedResult + textResponse.Value

		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)
		}
	}

	msg.Content = append(msg.Content,
		&types.ContentBlockMemberText{
			Value: combinedResult,
		},
	)

	return msg, nil
}
