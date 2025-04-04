package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"wox/common"
	"wox/setting"
	"wox/util"

	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
	"github.com/tidwall/gjson"
	"github.com/tmc/langchaingo/jsonschema"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

const groqBaseUrl = "https://api.groq.com/openai/v1"

type GroqProvider struct {
	connectContext setting.AIProvider
}

type GroqProviderStream struct {
	conversations []common.Conversation
	reader        io.Reader
}

func NewGroqProvider(ctx context.Context, connectContext setting.AIProvider) Provider {
	return &GroqProvider{connectContext: connectContext}
}

func (g *GroqProvider) ChatStream(ctx context.Context, model common.Model, conversations []common.Conversation, options common.ChatOptions) (ChatStream, error) {
	client, clientErr := openai.New(
		openai.WithModel(model.Name),
		openai.WithBaseURL(groqBaseUrl),
		openai.WithToken(g.connectContext.ApiKey),
		openai.WithHTTPClient(util.GetHTTPClient(ctx)),
	)
	if clientErr != nil {
		return nil, clientErr
	}

	buf := buffer.New(4 * 1024) // 4KB In memory Buffer
	r, w := nio.Pipe(buf)
	util.Go(ctx, "Groq chat stream", func() {
		_, err := client.GenerateContent(ctx, g.convertConversations(conversations), llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			w.Write(chunk)
			return nil
		}), llms.WithTools(g.convertTools(options.Tools)))
		if err != nil {
			w.CloseWithError(err)
		} else {
			w.Close()
		}
	})

	return &GroqProviderStream{conversations: conversations, reader: r}, nil
}

func (g *GroqProvider) convertTools(tools []common.MCPTool) []llms.Tool {
	convertedTools := make([]llms.Tool, len(tools))
	for i, tool := range tools {
		convertedTools[i] = llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"rationale": {
							Type:        jsonschema.String,
							Description: "The rationale for choosing this function call with these parameters",
						},
						"suggestions": {
							Type: jsonschema.Array,
							Items: &jsonschema.Definition{
								Type:        jsonschema.String,
								Description: "A suggested prompt",
							},
						},
					},
					Required: []string{"rationale", "suggestions"},
				},
			},
		}
	}
	return convertedTools
}

func (g *GroqProvider) Models(ctx context.Context) (models []common.Model, err error) {
	body, err := util.HttpGetWithHeaders(ctx, groqBaseUrl+"/models", map[string]string{
		"Authorization": "Bearer " + g.connectContext.ApiKey,
	})
	if err != nil {
		return nil, err
	}

	// response example
	//{
	//   "object": "list",
	//   "data": [
	//     {
	//       "id": "llama3-groq-70b-8192-tool-use-preview",
	//       "object": "model",
	//       "created": 1693721698,
	//       "owned_by": "Groq",
	//       "active": true,
	//       "context_window": 8192,
	//       "public_apps": null
	//     },
	//     {
	//       "id": "gemma2-9b-it",
	//       "object": "model",
	//       "created": 1693721698,
	//       "owned_by": "Google",
	//       "active": true,
	//       "context_window": 8192,
	//       "public_apps": null
	//     }
	//   ]
	// }

	// only return active models
	gjson.Get(string(body), "data").ForEach(func(key, value gjson.Result) bool {
		if value.Get("active").Bool() {
			models = append(models, common.Model{
				Name:     value.Get("id").String(),
				Provider: common.ProviderNameGroq,
			})
		}
		return true
	})

	return models, nil
}

func (g *GroqProvider) Ping(ctx context.Context) error {
	_, err := util.HttpGetWithHeaders(ctx, groqBaseUrl+"/models", map[string]string{
		"Authorization": "Bearer " + g.connectContext.ApiKey,
	})
	return err
}

func (g *GroqProvider) convertConversations(conversations []common.Conversation) (chatMessages []llms.MessageContent) {
	for _, conversation := range conversations {
		if conversation.Role == common.ConversationRoleUser {
			chatMessages = append(chatMessages, llms.TextParts(llms.ChatMessageTypeHuman, conversation.Text))
		}
		if conversation.Role == common.ConversationRoleAI {
			chatMessages = append(chatMessages, llms.TextParts(llms.ChatMessageTypeAI, conversation.Text))
		}
	}

	return chatMessages
}

func (s *GroqProviderStream) Receive(ctx context.Context) (string, error) {
	buf := make([]byte, 2048)
	n, err := s.reader.Read(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", io.EOF
		}
		return "", err
	}

	resp := string(buf[:n])
	util.GetLogger().Debug(util.NewTraceContext(), fmt.Sprintf("Groq: Send response: %s", resp))
	return resp, nil
}
