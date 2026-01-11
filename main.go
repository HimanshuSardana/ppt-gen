package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Topic struct {
	Title        string   `json:"title"`
	Introduction string   `json:"introduction"`
	TechStack    []string `json:"tech_stack"`
	Methodology  []string `json:"methodology"`
	FutureScope  []string `json:"future_scope"`
}

var TopicsSchema = GenerateSchema[Topic]()

func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

func askAI(prompt string) (string, error) {
	apiKey, present := os.LookupEnv("CEREBRAS_API_KEY")
	if !present {
		return "", fmt.Errorf("CEREBRAS_API_KEY environment variable not set")
	}

	client := openai.NewClient(
		option.WithBaseURL("https://api.cerebras.ai/v1"),
		option.WithAPIKey(apiKey),
		option.WithMaxRetries(5),
	)

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "ppt_content_generation",
		Description: openai.String("Given a topic, generate detailed PPT content including title, introduction, and tech stack."),
		Schema:      TopicsSchema,
		// Strict:      openai.Bool(true),
	}

	resp, err := client.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Model: "gpt-oss-120b",
			ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
					JSONSchema: schemaParam,
				},
			},
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(`You are a knowledgeable AI assistant specialized in creating engaging PowerPoint presentations. Given a specific topic, your task is to generate relevant content tailored to that subject area. Your goal is to produce informative slides that cover key points, supporting visuals, and concise explanations related to the provided topic.Key aspects of your role:1. Understand the given topic thoroughly.2. Organize content logically, starting from basic concepts and progressing to more advanced ideas.3. Incorporate appropriate visual aids such as diagrams, charts, or images to enhance comprehension.4. Use clear, concise language suitable for presentation delivery.5. Ensure consistency in formatting and style throughout the slides.6. Provide bullet points or short paragraphs for each main idea.7. Include relevant statistics or data when applicable.8. Conclude with a summary slide that recapitulates the key takeaways.Remember to focus on clarity, conciseness, and visual appeal to create an effective PowerPoint presentation.`),
				openai.UserMessage(prompt),
			},
		},
	)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

func main() {
	subject_name := "ttyDB (talk to your database): A jupyter-notebook style tool that allows querying data in natural language"

	prompt := fmt.Sprintf("Create detailed PPT content for the topic: %s. Include title, introduction, tech stack, methodology, and future scope. Do not include any emojis/special characters, it'll all be handled elsewhere", subject_name)
	response, err := askAI(prompt)
	if err != nil {
		log.Fatalf("Error asking AI: %v", err)
	}

	fmt.Println("AI Response:")
	// fmt.Println(response)

	UnmarshaledResponse := Topic{}
	err = json.Unmarshal([]byte(response), &UnmarshaledResponse)
	if err != nil {
		log.Fatalf("Error unmarshaling AI response: %v", err)
	}

	fmt.Printf("Unmarshaled Response: %+v\n", UnmarshaledResponse)

	data := Topic{
		Title:        UnmarshaledResponse.Title,
		Introduction: UnmarshaledResponse.Introduction,
		TechStack:    UnmarshaledResponse.TechStack,
		Methodology:  UnmarshaledResponse.Methodology,
		FutureScope:  UnmarshaledResponse.FutureScope,
	}

	os.WriteFile(subject_name+".json", []byte(response), 0644)

	// Write a typst file
	typstContent := fmt.Sprintf(`= Title: "%s"
	= Introduction
%s

	= Tech Stack
%s

	= Methodology
%s

	= Future Scope
%s
`, data.Title, data.Introduction, strings.Join((data.TechStack), "\n- "), strings.Join(data.Methodology, "\n- "), strings.Join(data.FutureScope, "\n- "))
	os.WriteFile(subject_name+".typ", []byte(typstContent), 0644)
}
