package aichat

import "time"

// 返回数据格式
type EventData struct {
	Name string
	Data string
}

type Stream struct {
	Data   string //一般是错误之间返回的信息
	Events chan *EventData
	Closed chan struct{}
}

// 配置相关
type Config struct {
	Auth     *Auth  //权限相关
	ProxyUrl string //代理url
}

// auth相关权限
type Auth struct {
	Token string //大多数会用到

	// claude相关
	organizationId string
}

/*****openai****/
// chatgpt
type ChatgptRequirement struct {
	Persona   string `json:"persona"`
	Token     string `json:"token"`
	Turnstile struct {
		Required bool   `json:"required"`
		Dx       string `json:"dx"`
	} `json:"turnstile"`
	Proofofwork struct {
		Required   bool   `json:"required"`
		Seed       string `json:"seed,omitempty"`
		Difficulty string `json:"difficulty"`
	} `json:"proofofwork,omitempty"`
	ForceLogin bool `json:"force_login,omitempty"`
}

type ChatgptCompletionRequest struct {
	Action                           string                       `json:"action"` //next
	ClientContextualInfo             *ChatgptClientContextualInfo `json:"client_contextual_info"`
	ConversationId                   string                       `json:"conversation_id,omitempty"`
	ConversationMode                 map[string]string            `json:"conversation_mode"`             //{"kind":"primary_assistant"}
	ConversationOrigin               any                          `json:"conversation_origin,omitempty"` //nil
	ForceParagen                     bool                         `json:"force_paragen"`                 //false
	ForceParagenModelSlug            string                       `json:"force_paragen_model_slug"`      //''
	ForceRateLimit                   bool                         `json:"force_rate_limit"`              //false
	HistoryAndTrainingDisabled       bool                         `json:"history_and_training_disabled"` //false
	Messages                         []*ChatgptRequestMessage     `json:"messages" binding:"required"`
	Model                            string                       `json:"model" binding:"required"`                       //auto
	ParagenCotSummaryDisplayOverride string                       `json:"paragen_cot_summary_display_override,omitempty"` //allow
	ParagenStreamTypeOverride        any                          `json:"paragen_stream_type_override,omitempty"`         //nil
	ParentMessageId                  string                       `json:"parent_message_id" binding:"required"`
	ResetRateLimits                  bool                         `json:"reset_rate_limits,omitempty"`   //false
	Suggestions                      []string                     `json:"suggestions"`                   //[]
	SupportedEncodings               []string                     `json:"supported_encodings,omitempty"` //["v1"]
	SupportsBuffering                bool                         `json:"supports_buffering,omitempty"`  //true
	SystemHints                      []any                        `json:"system_hints,omitempty"`        //[]
	Timezone                         string                       `json:"timezone,omitempty"`            //Asia/Shanghai
	TimezoneOffsetMin                float64                      `json:"timezone_offset_min,omitempty"` //-480
	WebsocketRequestId               string                       `json:"websocket_request_id"`
}

type ChatgptClientContextualInfo struct {
	IsDarkMode      bool `json:"is_dark_mode"`
	TimeSinceLoaded int  `json:"time_since_loaded"`
	PageHeight      int  `json:"page_height"`
	PageWidth       int  `json:"page_width"`
	PixelRatio      int  `json:"pixel_ratio"`
	ScreenHeight    int  `json:"screen_height"`
	ScreenWidth     int  `json:"screen_width"`
}

type ChatgptRequestMessage struct {
	ID         string          `json:"id" binding:"required"`
	Author     *ChatgptAuthor  `json:"author" binding:"required"` //{"role":"user"}
	Content    *ChatgptContent `json:"content" binding:"required"`
	CreateTime float64         `json:"create_time,omitempty"` //调用nowTimePay()
	// 默认{"serialization_metadata":{"custom_symbol_offsets":[]}}
	// 附件如下:
	// {"attachments":[{"id":"file-79q4G4iAyZQpPSNwqUmf7b","size":179129,
	// "name":"stock-photo-portrait-of-beautiful-long-haired-young-asian-office-woman-side-face-in-cream-suit-with-braces-on-2504794537.jpg",
	// "mime_type":"image/jpeg","width":1500,"height":1100}],"serialization_metadata":{"custom_symbol_offsets":[]}}
	Metadata map[string]any `json:"metadata,omitempty"`
}

type ChatgptAuthor struct {
	Role     string         `json:"role"`
	Name     string         `json:"name,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type ChatgptContent struct {
	ContentType string `json:"content_type" binding:"required"` //text multimodal_text
	// 附件类型时如:
	// {"content_type":"image_asset_pointer","asset_pointer":"file-service://file-79q4G4iAyZQpPSNwqUmf7b","size_bytes":179129,"width":1500,"height":1100}
	Parts []any `json:"parts" binding:"required"`
}

type ChatgptMetadata struct {
	Citations         []*Citation `json:"citations,omitempty"`
	MessageType       string      `json:"message_type,omitempty"`
	ModelSlug         string      `json:"model_slug,omitempty"`
	DefaultModelSlug  string      `json:"default_model_slug,omitempty"`
	ParentId          string      `json:"parent_id,omitempty"`
	ModelSwitcherDeny []any       `json:"model_switcher_deny,omitempty"`
}

type Citation struct {
	Metadata CitaMeta `json:"metadata"`
	StartIx  int      `json:"start_ix"`
	EndIx    int      `json:"end_ix"`
}

type CitaMeta struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type ChatgptCompletionResponse struct {
	V *ChatgptCompletionV `json:"v"`
	C int                 `json:"c"`
}

type ChatgptCompletionV struct {
	Message        *ChatgptCompletionVMessage `json:"message"`
	ConversationId string                     `json:"conversation_id"`
	Error          string                     `json:"error"`
}

type ChatgptCompletionVMessage struct {
	ID         string           `json:"id"`
	Author     *ChatgptAuthor   `json:"author"`
	CreateTime float64          `json:"create_time"`
	UpdateTime float64          `json:"update_time,omitempty"`
	Content    *ChatgptContent  `json:"content"`
	Status     string           `json:"status"`
	EndTurn    bool             `json:"end_turn,omitempty"`
	Weight     float64          `json:"weight"`
	Metadata   *ChatgptMetadata `json:"metadata"`
	Recipient  string           `json:"recipient"`
	Channel    string           `json:"channel,omitempty"`
}

type ChatgptRequest struct {
	MessageId       string   `json:"message_id,omitempty"`        //信息ID、标识
	ParentMessageId string   `json:"parent_message_id,omitempty"` //父信息ID、标识
	Files           []string `json:"files,omitempty"`             //附件内容base64
	ConversationId  string   `json:"conversation_id,omitempty"`   //会话ID
	ArkoseToken     string   `json:"arkose_token,omitempty"`
}

type ChatgptResponse struct {
	MessageId       string `json:"message_id"`                //本次会话信息ID
	ParentMessageId string `json:"parent_message_id"`         //对话返回的父信息ID,下次请求的parent_message_id用这个
	ConversationId  string `json:"conversation_id,omitempty"` //会话ID
}

// api或通用接口
type OpenaiChatCompletionRequest struct {
	Messages            []*ChatCompletionMessage `json:"messages"`
	Model               string                   `json:"model"`
	Store               bool                     `json:"store,omitempty"`                 //默认false
	ReasoningEffort     string                   `json:"reasoning_effort,omitempty"`      //默认medium
	Metadata            map[string]any           `json:"metadata,omitempty"`              //默认nil
	FrequencyPenalty    float64                  `json:"frequency_penalty,omitempty"`     //默认0
	LogitBias           map[string]int           `json:"logit_bias,omitempty"`            //默认nil
	Logprobs            bool                     `json:"logprobs,omitempty"`              //默认false
	TopLogprobs         int                      `json:"top_logprobs,omitempty"`          //0~20,logprobs设置为true才需设置
	MaxCompletionTokens int                      `json:"max_completion_tokens,omitempty"` //max_tokens作废
	N                   int                      `json:"n,omitempty"`                     //默认1
	Modalities          []string                 `json:"modalities,emitempty"`
	Prediction          map[string]any           `json:"prediction,omitempty"`
	Audio               map[string]any           `json:"audio,emitempty"`
	PresencePenalty     float64                  `json:"presence_penalty,omitempty"` //默认0
	ResponseFormat      map[string]string        `json:"response_format,omitempty"`
	Seed                int                      `json:"seed,omitempty"`
	ServiceTier         string                   `json:"service_tier,omitempty"` //默认auto
	Stop                any                      `json:"stop,omitempty"`
	Stream              bool                     `json:"stream,omitempty"`
	StreamOptions       map[string]bool          `json:"stream_options,omitempty"` //stream设置为true时有效
	Temperature         float64                  `json:"temperature,omitempty"`    //默认1
	TopP                float64                  `json:"top_p,omitempty"`          //默认1
	Tools               []Tool                   `json:"tools,omitempty"`
	ToolChoice          any                      `json:"tool_choice,omitempty"`
	ParallelToolCalls   bool                     `json:"parallel_tool_calls,omitempty"` //默认true
	User                string                   `json:"user,omitempty"`

	// chatgpt专用参数
	ChatgptExt *ChatgptRequest `json:"chatgpt,omitempty"`

	// claude ai专用参数
	ClaudeExt *ClaudeExtRequest `json:"claude,omitempty"`
}

type ChatCompletionResponse struct {
	ID                string                  `json:"id"`
	Choices           []*ChatCompletionChoice `json:"choices"`
	Created           int64                   `json:"created"`
	Model             string                  `json:"model"`
	ServiceTier       string                  `json:"service_tier,omitempty"`
	SystemFingerprint string                  `json:"system_fingerprint"`
	Object            string                  `json:"object"`
	Usage             *Usage                  `json:"usage,omitempty"`

	// 自定义chatgpt额外信息
	Chatgpt *ChatgptResponse `json:"chatgpt,omitempty"`

	// 自定义claude额外信息
	Claude *ClaudeExtResponse `json:"claude,omitempty"`
}

type ChatCompletionRequest struct {
	Source string `json:"source,omitempty"` //原始json字符串请求

	// openai
	Openai  *OpenaiChatCompletionRequest //api或者通用
	Chatgpt *ChatgptCompletionRequest    //chatgpt

	// claude
	Anthropic *AnthropicChatRequest //anthropic api
	Claude    *ClaudeChatRequest    //claude ai web chat
}

func (c *OpenaiChatCompletionRequest) ParsePromptText() string {
	prompt := ""
	for k := range c.Messages {
		message := c.Messages[k]
		if message.Role == "user" {
			if message.Contents == nil {
				prompt = message.Content
			}
		}
	}
	return prompt
}

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
	// Assistant message
	Refusal   string            `json:"refusal,omitempty"`
	Audio     map[string]string `json:"audio,omitempty"`
	ToolCalls []*ToolCall       `json:"tool_calls,omitempty"`
	// Tool message
	ToolCallID string `json:"tool_call_id,omitempty"`
	Contents   []*ChatMessagePart
}

func (m *ChatCompletionMessage) MarshalJSON() ([]byte, error) {
	if m.Content != "" && m.Contents != nil {
		return nil, errContentFieldsMisused
	}
	if len(m.Contents) > 0 {
		msg := struct {
			Role       string             `json:"role"`
			Content    string             `json:"-"`
			Name       string             `json:"name,omitempty"`
			Refusal    string             `json:"refusal,omitempty"`
			Audio      map[string]string  `json:"audio,omitempty"`
			ToolCalls  []*ToolCall        `json:"tool_calls,omitempty"`
			ToolCallID string             `json:"tool_call_id,omitempty"`
			Contents   []*ChatMessagePart `json:"content,omitempty"`
		}(*m)
		return Json.Marshal(msg)
	}
	msg := struct {
		Role       string             `json:"role"`
		Content    string             `json:"content"`
		Name       string             `json:"name,omitempty"`
		Refusal    string             `json:"refusal,omitempty"`
		Audio      map[string]string  `json:"audio,omitempty"`
		ToolCalls  []*ToolCall        `json:"tool_calls,omitempty"`
		ToolCallID string             `json:"tool_call_id,omitempty"`
		Contents   []*ChatMessagePart `json:"-"`
	}(*m)
	return Json.Marshal(msg)
}

func (m *ChatCompletionMessage) UnmarshalJSON(bs []byte) error {
	msg := struct {
		Role       string            `json:"role"`
		Content    string            `json:"content"`
		Name       string            `json:"name,omitempty"`
		Refusal    string            `json:"refusal,omitempty"`
		Audio      map[string]string `json:"audio,omitempty"`
		ToolCalls  []*ToolCall       `json:"tool_calls,omitempty"`
		ToolCallID string            `json:"tool_call_id,omitempty"`
		Contents   []*ChatMessagePart
	}{}
	if err := Json.Unmarshal(bs, &msg); err == nil {
		*m = ChatCompletionMessage(msg)
		return nil
	}
	multiMsg := struct {
		Role       string `json:"role"`
		Content    string
		Name       string             `json:"name,omitempty"`
		Refusal    string             `json:"refusal,omitempty"`
		Audio      map[string]string  `json:"audio,omitempty"`
		ToolCalls  []*ToolCall        `json:"tool_calls,omitempty"`
		ToolCallID string             `json:"tool_call_id,omitempty"`
		Contents   []*ChatMessagePart `json:"content"`
	}{}
	if err := Json.Unmarshal(bs, &multiMsg); err != nil {
		return err
	}
	*m = ChatCompletionMessage(multiMsg)
	return nil
}

type ChatMessagePart struct {
	Type       string                 `json:"type"`
	Text       string                 `json:"text,omitempty"`
	ImageUrl   *ChatMessageImageUrl   `json:"image_url,omitempty"`
	InputAudio *ChatMessageInputAudio `json:"input_audio,omitempty"`
}

type ChatMessageImageUrl struct {
	Url    string `json:"url"` //url或base64数据
	Detail string `json:"detail,omitempty"`
}

type ChatMessageInputAudio struct {
	Data   string `json:"data"`   //base64
	Format string `json:"format"` //格式wav或mp3
}

type ToolCall struct {
	ID       string        `json:"id"`
	Type     string        `json:"type"`
	Function *FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type Tool struct {
	Type     string              `json:"type"`
	Function *FunctionDefinition `json:"function"`
}

type FunctionDefinition struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty"`
	Strict      bool              `json:"strict,omitempty"` //默认false
}

type ChatCompletionChoice struct {
	Index        int                    `json:"index"`
	Message      *ChatCompletionMessage `json:"message,omitempty"`
	Delta        *ChatCompletionMessage `json:"delta,omitempty"`
	LogProbs     *LogProbs              `json:"logprobs"`
	FinishReason string                 `json:"finish_reason"`
}

type LogProbs struct {
	Content []*LogProbContent `json:"content,omitempty"`
	Refusal []*LogProbContent `json:"refusal,omitempty"`
}

type LogProbContent struct {
	LogProb
	TopLogprobs []LogProb `json:"top_logprobs"`
}

type LogProb struct {
	Token   string  `json:"token"`
	Logprob float64 `json:"logprob"`
	Bytes   []byte  `json:"bytes,omitempty"`
}

type Usage struct {
	CompletionTokens        int            `json:"completion_tokens"`
	PromptTokens            int            `json:"prompt_tokens"`
	TotalTokens             int            `json:"total_tokens"`
	CompletionTokensDetails map[string]int `json:"completion_tokens_details,omitempty"`
	PromptTokensDetails     map[string]int `json:"prompt_tokens_details,omitempty"`
}

/*****openai end****/

/*****claude*****/
// 创建conversation请求
type ClaudeCreateConversationRequest struct {
	Uuid                           string `json:"uuid"`
	Name                           string `json:"name"`
	IncludeConversationPreferences bool   `json:"include_conversation_preferences"` //true
}

type ClaudeConversation struct {
	ID                   string         `json:"uuid"`
	Name                 string         `json:"name"`
	Summary              string         `json:"summary"`
	Model                any            `json:"model"`
	CreatedAt            string         `json:"created_at"`
	UpdatedAt            string         `json:"updated_at"`
	Settings             map[string]any `json:"settings"`
	IsStarred            bool           `json:"is_starred"`
	ProjectId            string         `json:"project_uuid,omitempty"`
	CurrentLeafMessageId string         `json:"current_leaf_message_uuid,omitempty"`
}

type ClaudeExtRequest struct {
	ParentMessageId string `json:"parent_message_id,omitempty"` //父信息ID、标识
	Attachments     []any  `json:"attachments,omitempty"`       //附件
	Files           []any  `json:"files,omitempty"`             //文件
	ConversationId  string `json:"conversation_id,omitempty"`   //会话ID
	OrganizationId  string `json:"organization_id,omitempty"`   //组织ID
}

type ClaudeExtResponse struct {
	ParentId       string `json:"parent_id"`       //对话返回的父信息ID,下次请求的parent_message_id用这个
	ConversationId string `json:"conversation_id"` //会话ID
}

// web chat请求
type ClaudeChatRequest struct {
	Prompt             string                     `json:"prompt"`
	ParentMessageUuid  string                     `json:"parent_message_uuid"` //初始00000000-0000-4000-8000-000000000000
	Timezone           string                     `json:"timezone"`
	PersonalizedStyles []*ClaudePersonalizedStyle `json:"personalized_styles"`
	Attachments        []any                      `json:"attachments,omitempty"`
	Files              []any                      `json:"files,omitempty"`
	SyncSources        []any                      `json:"sync_sources,omitempty"`
	RenderingMode      string                     `json:"rendering_mode"` //可能取值messages

	ClaudeExt *ClaudeExtRequest `json:"claude,omitempty"` //自定义额外请求
}

// 页面请求返回
type ClaudeChatResponse struct {
	Type         string              `json:"type"` //message_start、content_block_delta等
	Message      *ClaudeChatMessage  `json:"message,omitempty"`
	Index        int                 `json:"index,omitempty"`
	Delta        *ClaudeChatDelta    `json:"delta,omitempty"`
	MessageLimit *ClaudeMessageLimit `json:"message_limit,omitempty"`
}

type ClaudeChatMessage struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	Model        string `json:"model"`
	ParentUuid   string `json:"parent_uuid"`
	Uuid         string `json:"uuid"`
	Content      []any  `json:"content"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

type ClaudeChatDelta struct {
	Type string `json:"type"` //text_delta、input_json_delta
	Text string `json:"text,omitempty"`
	// type为input_json_delta时才有
	PartialJson string `json:"partial_json,omitempty"` //json数据
	// type为message_delta时才有
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

type ClaudeMessageLimit struct {
	Type          string `json:"type"`
	ResetsAt      string `json:"resetsAt,omitempty"`
	Remaining     string `json:"remaining,omitempty"`
	PerModelLimit any    `json:"perModelLimit,omitempty"`
}

type ClaudePersonalizedStyle struct {
	Name      string `json:"name"`      //"Normal"
	Prompt    string `json:"prompt"`    //"Normal"
	Summary   string `json:"summary"`   //"Default responses from Claude"
	IsDefault bool   `json:"isDefault"` //true
	Type      string `json:"type"`      //"default"
	Key       string `json:"key"`       //"Default"
}

type ClaudeOrganization struct {
	ID                       int64          `json:"id"`
	Uuid                     string         `json:"uuid"`
	Name                     string         `json:"name"`
	Settings                 map[string]any `json:"settings"`
	Capabilities             []string       `json:"capabilities"`
	ParentOrganizationUuid   string         `json:"parent_organization_uuid,omitempty"`
	RateLimitTier            string         `json:"rate_limit_tier"`
	BillingType              string         `json:"billing_type,omitempty"`
	FreeCreditsStatus        string         `json:"free_credits_status,omitempty"`
	DataRetention            string         `json:"data_retention,omitempty"`
	ApiDisabledReason        string         `json:"api_disabled_reason,omitempty"`
	ApiDisabledUntil         any            `json:"api_disabled_until,omitempty"`
	BillableUsagePausedUntil any            `json:"billable_usage_paused_until"`
	RavenType                any            `json:"raven_type,omitempty"`
	CreatedAt                string         `json:"created_at"`
	UpdatedAt                string         `json:"updated_at"`
	ActiveFlags              []any          `json:"active_flags"`
	DataRetentionPeriods     any            `json:"data_retention_periods,omitempty"`
}

// api请求
type AnthropicChatRequest struct {
	Model         string                   `json:"model"`
	Messages      []*AnthropicChatMessage  `json:"messages"`
	MaxTokens     int                      `json:"max_tokens"`
	Metadata      *AnthropicChatMetadata   `json:"metadata,omitempty"`
	StopSequences []string                 `json:"stop_sequences,omitempty"`
	Stream        bool                     `json:"stream,omitempty"`
	Temperature   float64                  `json:"temperature,omitempty"` //0 < x < 1,默认1.0
	ToolChoice    *AnthropicChatToolChoice `json:"tool_choice,omitempty"`
	Tools         []*AnthropicChatTool     `json:"tools,omitempty"`
	TopK          int                      `json:"top_k,omitempty"` //>0
	TopP          float64                  `json:"top_p,omitempty"` //0 < x < 1

	ClaudeExt *ClaudeExtRequest `json:"claude,omitempty"` //自定义额外信息
}

func (a *AnthropicChatRequest) ParsePromptText() string {
	prompt := ""
	for k := range a.Messages {
		message := a.Messages[k]
		if message.Role == "user" {
			if message.Contents == nil {
				prompt = message.Content
			}
		}
	}
	return prompt
}

// api返回
type AnthropicChatResponse struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"` //可能取值message
	Role         string                  `json:"role"` //如assistant
	Content      []*AnthropicChatContent `json:"content"`
	Model        string                  `json:"model"`
	StopReason   string                  `json:"stop_reason,omitempty"` //end_turn max_tokens stop_sequence tool_use
	StopSequence string                  `json:"stop_sequence,omitempty"`
	Usage        *AnthropicChatUsage     `json:"usage"`
}

func (a *AnthropicChatResponse) ToChatCompletionResponse() *ChatCompletionResponse {
	var ccc []*ChatCompletionChoice
	if len(a.Content) > 0 {
		for k := range a.Content {
			content := a.Content[k]
			ccc = append(ccc, &ChatCompletionChoice{
				Index: 0,
				Message: &ChatCompletionMessage{
					Role:    "assistant",
					Content: content.Text,
				},
				FinishReason: "stop",
			})
		}
	}
	return &ChatCompletionResponse{
		ID:                a.ID,
		Choices:           ccc,
		Created:           time.Now().Unix(),
		Model:             a.Model,
		SystemFingerprint: a.ID,
		Object:            "chat.completion",
		Usage: &Usage{
			CompletionTokens: a.Usage.OutputTokens,
			PromptTokens:     a.Usage.InputTokens,
			TotalTokens:      a.Usage.OutputTokens + a.Usage.InputTokens,
		},
	}
}

// api stream返回
type AnthropicChatStreamResponse struct {
	// 类型
	// ping、content_block_start、error
	// content_block_delta、message_start
	// content_block_stop、message_delta
	// message_stop
	Type string `json:"type"`

	Index int `json:"index,omitempty"`

	// type为content_block_start
	ContentBlock *ClaudeChatDelta `json:"content_block,omitempty"`

	// type为message_start时
	Message *AnthropicChatResponse `json:"message,omitempty"`

	// type为message_delta、content_block_delta(主要信息)
	Delta *ClaudeChatDelta `json:"delta,omitempty"`

	// type为message_delta时
	Usage *AnthropicChatUsage `json:"usage,omitempty"`

	// error错误
	Error *AnthropicChatError `json:"error,omitempty"`

	// 自定义claude额外信息
	Claude *ClaudeExtResponse `json:"claude,omitempty"`
}

// func (a *AnthropicChatStreamResponse) ToChatCompletionResponse(model string) *ChatCompletionResponse {
// 	var ccc []*ChatCompletionChoice
// 	if len(a.Content) > 0 {
// 		for k := range a.Content {
// 			content := a.Content[k]
// 			ccc = append(ccc, &ChatCompletionChoice{
// 				Index: 0,
// 				Message: &ChatCompletionMessage{
// 					Role:    "assistant",
// 					Content: content.Text,
// 				},
// 				FinishReason: "stop",
// 			})
// 		}
// 	}
// 	return &ChatCompletionResponse{
// 		ID:                a.ID,
// 		Choices:           ccc,
// 		Created:           time.Now().Unix(),
// 		Model:             model,
// 		SystemFingerprint: a.ID,
// 		Object:            "chat.completion",
// 	}
// }

type AnthropicChatError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type AnthropicChatMessage struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	Contents []*AnthropicChatContent
}

func (c *AnthropicChatMessage) MarshalJSON() ([]byte, error) {
	if c.Content != "" && c.Contents != nil {
		return nil, errContentFieldsMisused
	}
	if len(c.Contents) > 0 {
		msg := struct {
			Role     string                  `json:"role"`
			Content  string                  `json:"-"`
			Contents []*AnthropicChatContent `json:"content"`
		}(*c)
		return Json.Marshal(msg)
	}
	msg := struct {
		Role     string                  `json:"role"`
		Content  string                  `json:"content"`
		Contents []*AnthropicChatContent `json:"-"`
	}(*c)
	return Json.Marshal(msg)
}

func (c *AnthropicChatMessage) UnmarshalJSON(bs []byte) error {
	msg := struct {
		Role     string `json:"role"`
		Content  string `json:"content"`
		Contents []*AnthropicChatContent
	}{}
	if err := Json.Unmarshal(bs, &msg); err == nil {
		*c = AnthropicChatMessage(msg)
		return nil
	}
	multiMsg := struct {
		Role     string `json:"role"`
		Content  string
		Contents []*AnthropicChatContent `json:"content"`
	}{}
	if err := Json.Unmarshal(bs, &multiMsg); err != nil {
		return err
	}
	*c = AnthropicChatMessage(multiMsg)
	return nil
}

type AnthropicChatContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`

	// 图片、文件等
	Source *AnthropicChatSource `json:"source,omitempty"`

	// tool use
	ID    string         `json:"id,omitempty"`
	Name  string         `json:"name,omitempty"`
	Input map[string]any `json:"input,omitempty"`
}

type AnthropicChatSource struct {
	Type      string `json:"type"`       //base64
	MediaType string `json:"media_type"` //image/jpeg
	Data      string `json:"data"`       //数据base64
}

type AnthropicChatMetadata struct {
	UserId string `json:"user_id,omitempty"` //标识,如uuid或用户邮箱等等
}

type AnthropicChatToolChoice struct {
	Type                   string `json:"type"`
	Name                   string `json:"name,omitempty"`
	DisableParallelToolUse bool   `json:"disable_parallel_tool_use"`
}

type AnthropicChatTool struct {
	Type            string         `json:"type"`
	Description     string         `json:"description,omitempty"`
	Name            string         `json:"name,omitempty"`
	InputSchema     map[string]any `json:"input_schema,omitempty"`
	CacheControl    map[string]any `json:"cache_control,omitempty"`
	DisplayHeightPx int            `json:"display_height_px,omitempty"`
	DisplayWidthPx  int            `json:"display_width_px,omitempty"`
	DisplayNumber   int            `json:"display_number,omitempty"`
}

type AnthropicChatUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}

/*****claude end*****/
