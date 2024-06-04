package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
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

//const modelID = "anthropic.claude-3-haiku-20240307-v1:0"

func main() {
	reader := bufio.NewReader(os.Stdin)

	converseInput := &bedrockruntime.ConverseInput{
		ModelId: aws.String(modelID),
	}

	for {
		fmt.Print("\nChoose your message type - Text (enter 1) or Image (enter 2): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		userMsg := types.Message{
			Role: types.ConversationRoleUser,
		}

		if input == "1" {

			fmt.Print("\nEnter your message: ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)

			userMsg.Content = append(userMsg.Content, &types.ContentBlockMemberText{
				Value: text,
			})
		} else if input == "2" {

			for {
				fmt.Print("\nEnter the image source (local path or url): ")
				path, _ := reader.ReadString('\n')
				path = strings.TrimSpace(path)

				imageContents, err := readImage(path)
				if err != nil {
					log.Fatal(err)
				}

				userMsg.Content = append(userMsg.Content, &types.ContentBlockMemberImage{
					Value: types.ImageBlock{
						Format: types.ImageFormatJpeg,
						Source: &types.ImageSourceMemberBytes{
							Value: imageContents,
						},
					},
				})

				fmt.Print("\nWould you like to add more images? enter yes or no: ")
				yesOrNo, _ := reader.ReadString('\n')
				yesOrNo = strings.TrimSpace(yesOrNo)

				if yesOrNo == "no" {
					fmt.Print("\nWhat would you like to ask about the image(s)? : ")
					q, _ := reader.ReadString('\n')
					q = strings.TrimSpace(q)

					userMsg.Content = append(userMsg.Content, &types.ContentBlockMemberText{
						Value: q,
					})

					break
				} else if yesOrNo == "yes" {
					continue
				} else {
					log.Fatal("invalid option. enter yes or no. start over again")
				}
			}

		} else {
			log.Fatal("invalid option. enter 1 or 2. start over again")
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

func readImage(source string) ([]byte, error) {

	var imageBytes []byte

	if strings.Contains(source, "http") {
		resp, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		imageBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		//assume it's local
		var err error
		imageBytes, err = os.ReadFile(source)
		if err != nil {
			return nil, err
		}
	}

	return imageBytes, nil
}
