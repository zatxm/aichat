package aichat

type EventData struct {
	Name string
	Data string
}

type Stream struct {
	Events chan *EventData
	Closed chan struct{}
}

// 通用请求
type RequestChat struct {
	MessageId       string   `json:"message_id,omitempty"`        //信息ID、标识
	ParentMessageId string   `json:"parent_message_id,omitempty"` //父信息ID、标识
	Model           string   `json:"model,omitempty"`             //对话模型
	Message         string   `json:"message" binding:"required"`  //请求信息
	Files           []string `json:"files,omitempty"`             //附件内容base64
	ConversationId  string   `json:"conversation_id,omitempty"`   //会话ID
	Ext             any      `json:"ext,omitempty"`               //额外信息
}

/*****chatgpt****/
type ChatRequirement struct {
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
	ID         string                 `json:"id" binding:"required"`
	Author     map[string]string      `json:"author" binding:"required"` //{"role":"user"}
	Content    *ChatgptRequestContent `json:"content" binding:"required"`
	CreateTime float64                `json:"create_time,omitempty"` //调用nowTimePay()
	// 默认{"serialization_metadata":{"custom_symbol_offsets":[]}}
	// 附件如下:
	// {"attachments":[{"id":"file-79q4G4iAyZQpPSNwqUmf7b","size":179129,
	// "name":"stock-photo-portrait-of-beautiful-long-haired-young-asian-office-woman-side-face-in-cream-suit-with-braces-on-2504794537.jpg",
	// "mime_type":"image/jpeg","width":1500,"height":1100}],"serialization_metadata":{"custom_symbol_offsets":[]}}
	Metadata map[string]any `json:"metadata,omitempty"`
}

type ChatgptRequestContent struct {
	ContentType string `json:"content_type" binding:"required"` //text multimodal_text
	// 附件类型时如:
	// {"content_type":"image_asset_pointer","asset_pointer":"file-service://file-79q4G4iAyZQpPSNwqUmf7b","size_bytes":179129,"width":1500,"height":1100}
	Parts []any `json:"parts" binding:"required"`
}
