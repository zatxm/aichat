package aichat

type EventData struct {
	Name string
	Data string
}

type Stream struct {
	Data   string //一般是错误之间返回的信息
	Events chan *EventData
	Closed chan struct{}
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
	Index           string   `json:"index,omitempty"` //用户定义逻辑ID
}

type ChatgptResponse struct {
	MessageId       string `json:"message_id"`                //本次会话信息ID
	ParentMessageId string `json:"parent_message_id"`         //对话返回的父信息ID,下次请求的parent_message_id用这个
	ConversationId  string `json:"conversation_id,omitempty"` //会话ID
	Index           string `json:"index,omitempty"`
}

// openai api
type ChatCompletionRequest struct {
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

	Source string `json:"source,omitempty"` //原始json字符串请求

	Provider string `json:"provider,omitempty"` //会话类型,非openai api参数

	// chatgpt专用参数
	Chatgpt *ChatgptRequest `json:"chatgpt,omitempty"`
}

func (c *ChatCompletionRequest) ParsePromptText() string {
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

type ChatCompletionResponse struct {
	ID                string                  `json:"id"`
	Choices           []*ChatCompletionChoice `json:"choices"`
	Created           int64                   `json:"created"`
	Model             string                  `json:"model"`
	ServiceTier       string                  `json:"service_tier,omitempty"`
	SystemFingerprint string                  `json:"system_fingerprint"`
	Object            string                  `json:"object"`
	Usage             *Usage                  `json:"usage,omitempty"`

	// chatgpt额外信息
	Chatgpt *ChatgptResponse `json:"chatgpt,omitempty"`
}

type ChatCompletionChoice struct {
	Delta        *ChatCompletionMessage `json:"delta,omitempty"`
	LogProbs     *LogProbs              `json:"logprobs"`
	FinishReason string                 `json:"finish_reason"`
	Index        int                    `json:"index"`
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
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

/*****openai end****/
