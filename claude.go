package aichat

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/google/uuid"
	mreq "github.com/imroc/req/v3"
	"go.uber.org/zap"
)

type incomparable [0]func()

type brReader struct {
	_    incomparable
	body io.ReadCloser
	zr   *brotli.Reader
	zerr error
}

func (br *brReader) Read(p []byte) (n int, err error) {
	if br.zerr != nil {
		return 0, br.zerr
	}
	if br.zr == nil {
		br.zr = brotli.NewReader(br.body)
	}
	return br.zr.Read(p)
}

func (br *brReader) Close() error {
	return br.body.Close()
}

type claude struct {
	config *Config
	client *mreq.Client
}

func NewClaude(cfg *Config) AiCommon {
	client := mreq.C().SetUserAgent(userAgent).ImpersonateChrome()
	if cfg.ProxyUrl != "" {
		client.SetProxyURL(cfg.ProxyUrl)
	}
	// 返回br解码
	client.OnAfterResponse(func(client *mreq.Client, resp *mreq.Response) error {
		if resp.Err != nil {
			return nil
		}

		if resp.Header.Get("Content-Encoding") == "br" {
			resp.Body = &brReader{
				body: resp.Body,
			}
		}

		return nil
	})
	if cfg.Auth == nil {
		cfg.Auth = &Auth{}
	}
	client.SetCommonHeaders(map[string]string{
		"accept-encoding": acceptEncoding,
		"cookie":          "sessionKey=" + cfg.Auth.Token,
		"content-type":    contentTypeJson,
		"origin":          "https://claude.ai",
		"referer":         "https://claude.ai/chats",
	})
	return &claude{config: cfg, client: client}
}

func (c *claude) SetProxy(proxyUrl string) {
	c.config.ProxyUrl = proxyUrl
	if proxyUrl != "" {
		c.client.SetProxyURL(proxyUrl)
	}
}

func (c *claude) SetAuth(auth *Auth) {
	c.config.Auth = auth
	c.client.SetCommonHeader("Cookie", "sessionKey="+auth.Token)
}

func (c *claude) Chat(rc *ChatCompletionRequest) (*Stream, error) {
	ccr, err := c.toClaudeChatRequest(rc)
	if err != nil {
		return nil, err
	}
	return c.goChatWeb(ccr)
}

func (c *claude) ChatApi(rc *ChatCompletionRequest) (*Stream, error) {
	return c.doApiOrigin(rc)
}

func (c *claude) ChatToApi(rc *ChatCompletionRequest) (*Stream, error) {
	ccr, err := c.toClaudeChatRequest(rc)
	if err != nil {
		return nil, err
	}
	return c.goChatWebToApi(ccr)
}

func (c *claude) ApiCrossChatToApi(rc *ChatCompletionRequest) (*Stream, error) {
	ccr, err := c.apiToClaudeChatRequest(rc)
	if err != nil {
		return nil, err
	}
	return c.goChatWebToApi(ccr)
}

func (c *claude) ChatToOpenaiApi(rc *ChatCompletionRequest) (*Stream, error) {
	ccr, err := c.toClaudeChatRequest(rc)
	if err != nil {
		return nil, err
	}
	return c.goChatWebToOpenaiApi(ccr)
}

func (c *claude) ApiToOpenaiApi(rc *ChatCompletionRequest) (*Stream, error) {
	return c.doApiOpenaiApi(rc)
}

func (c *claude) ApiCrossChatToOpenaiApi(rc *ChatCompletionRequest) (*Stream, error) {
	ccr, err := c.apiToClaudeChatRequest(rc)
	if err != nil {
		return nil, err
	}
	return c.goChatWebToOpenaiApi(ccr)
}

func (c *claude) CommonChatToOpenaiApi(rc *ChatCompletionRequest) (*Stream, error) {
	ccr, err := c.commonToClaudeChatRequest(rc)
	if err != nil {
		return nil, err
	}
	return c.goChatWebToOpenaiApi(ccr)
}

// 转成web chat的ClaudeChatRequest请求体
func (c *claude) toClaudeChatRequest(rc *ChatCompletionRequest) (*ClaudeChatRequest, error) {
	if rc.Source == "" && rc.Claude == nil {
		return nil, errors.New("params error")
	}
	ccr := &ClaudeChatRequest{}
	if rc.Source != "" {
		err := Json.UnmarshalFromString(rc.Source, &ccr)
		if err != nil {
			return nil, err
		}
	} else {
		ccr = rc.Claude
	}
	return ccr, nil
}

// 转成api的AnthropicChatRequest请求体
func (c *claude) toAnthropicChatRequest(rc *ChatCompletionRequest) (*AnthropicChatRequest, error) {
	if rc.Source == "" && rc.Anthropic == nil {
		return nil, errors.New("params error")
	}
	acr := &AnthropicChatRequest{}
	if rc.Source != "" {
		err := Json.UnmarshalFromString(rc.Source, &acr)
		if err != nil {
			return nil, err
		}
		rc.Anthropic = &AnthropicChatRequest{Model: acr.Model}
	} else {
		acr = rc.Anthropic
	}
	return acr, nil
}

// api转成web chat的ClaudeChatRequest请求体
func (c *claude) apiToClaudeChatRequest(rc *ChatCompletionRequest) (*ClaudeChatRequest, error) {
	arc, err := c.toAnthropicChatRequest(rc)
	if err != nil {
		return nil, err
	}
	prompt := arc.ParsePromptText()
	if prompt == "" {
		return nil, errors.New("request error:message empty")
	}
	if arc.ClaudeExt == nil {
		arc.ClaudeExt = &ClaudeExtRequest{}
	}
	parentMsgId := arc.ClaudeExt.ParentMessageId
	if parentMsgId == "" {
		parentMsgId = "00000000-0000-4000-8000-000000000000"
	}
	ccr := &ClaudeChatRequest{
		Prompt:             prompt,
		ParentMessageUuid:  parentMsgId,
		Timezone:           "Asia/Shanghai",
		PersonalizedStyles: c.generateClaudePersonalizedStyle(),
		RenderingMode:      "messages",
		ClaudeExt: &ClaudeExtRequest{
			ConversationId: arc.ClaudeExt.ConversationId,
			OrganizationId: arc.ClaudeExt.OrganizationId},
	}
	return ccr, nil
}

// api请求
func (c *claude) doApiRequest(acr *AnthropicChatRequest) (*mreq.Response, error) {
	reqByte, err := Json.Marshal(acr)
	if err != nil {
		return nil, err
	}
	// 设置请求端
	client := mreq.C().SetCommonHeaders(
		map[string]string{
			"anthropic-version": anthropicVersion,
			"content-type":      contentTypeJson,
			"x-api-key":         c.config.Auth.Token,
		})
	if c.config.ProxyUrl != "" {
		client.SetProxyURL(c.config.ProxyUrl)
	}
	rq := client.R()
	if acr.Stream {
		rq.SetHeader("accept", "text/event-stream")
	}

	return rq.SetBodyBytes(reqByte).Post(anthropicApiUrl + "/v1/messages")
}

// 完整api请求、返回
func (c *claude) doApiOrigin(rc *ChatCompletionRequest) (*Stream, error) {
	acr, err := c.toAnthropicChatRequest(rc)
	if err != nil {
		return nil, err
	}

	resp, err := c.doApiRequest(acr)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("request chat return http status error")
	}

	// 处理返回
	if !acr.Stream {
		b, err := readAllToString(resp.Body)
		if err != nil {
			return nil, err
		}
		return &Stream{Data: b}, nil
	}

	// 处理stream流
	return originStream(resp)
}

// api请求转成openai api返回格式
func (c *claude) doApiOpenaiApi(rc *ChatCompletionRequest) (*Stream, error) {
	acr, err := c.toAnthropicChatRequest(rc)
	if err != nil {
		return nil, err
	}

	resp, err := c.doApiRequest(acr)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("request chat return http status error")
	}

	// 处理返回
	if !acr.Stream {
		var acrRes AnthropicChatResponse
		err := Json.NewDecoder(resp.Body).Decode(&acrRes)
		if err != nil {
			return nil, err
		}
		ccrRes := acrRes.ToChatCompletionResponse()
		b, err := Json.MarshalToString(ccrRes)
		if err != nil {
			return nil, err
		}
		return &Stream{Data: b}, nil
	}

	// 处理br stream流
	resMsgId, model := "", ""
	stream := &Stream{
		Events: make(chan *EventData),
		Closed: make(chan struct{}),
	}
	go func() {
		var ir io.Reader
		if resp.Header.Get("Content-Encoding") == "br" {
			ir = brotli.NewReader(resp.Body)
		} else {
			ir = resp.Body
		}
		reader := bufio.NewReader(ir)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if strings.HasPrefix(line, "data:") {
				raw := line[6:]
				raw = strings.TrimSuffix(raw, "\n")
				var chatRes AnthropicChatStreamResponse
				err = Json.UnmarshalFromString(raw, &chatRes)
				if err != nil {
					continue
				}
				switch chatRes.Type {
				case "message_start":
					resMsgId = chatRes.Message.ID
					model = chatRes.Message.Model
				case "content_block_delta":
					var choices []*ChatCompletionChoice
					choices = append(choices, &ChatCompletionChoice{
						Delta: &ChatCompletionMessage{
							Role:    "assistant",
							Content: chatRes.Delta.Text,
						},
						Index: chatRes.Index,
					})
					outRes := &ChatCompletionResponse{
						ID:      resMsgId,
						Choices: choices,
						Created: time.Now().Unix(),
						Model:   model,
						Object:  "chat.completion.chunk",
					}
					outJson, _ := Json.MarshalToString(outRes)
					stream.Events <- &EventData{Name: "", Data: "data: " + outJson + "\n\n"}
				}
			}
		}
		stream.Events <- &EventData{Name: "", Data: "data: [DONE]\n\n"}
	}()
	return stream, nil
}

func (c *claude) doChatWebRequest(ccr *ClaudeChatRequest) (*mreq.Response, error, string) {
	conversationId, organizationId := "", ""
	if ccr.ClaudeExt != nil {
		conversationId = ccr.ClaudeExt.ConversationId
		organizationId = ccr.ClaudeExt.OrganizationId
		ccr.ClaudeExt = nil
	}
	ccrByte, err := Json.Marshal(ccr)
	if err != nil {
		return nil, err, ""
	}
	if organizationId == "" {
		organizationId, err = c.generateOrganizationId(false)
		if err != nil {
			return nil, err, ""
		}
		if organizationId == "" {
			return nil, errors.New("request chat parse organization_id error"), ""
		}
	}
	if conversationId == "" {
		// 新会话
		createConversationUrl := "https://claude.ai/api/organizations/" + organizationId + "/chat_conversations"
		rq := &ClaudeCreateConversationRequest{
			Uuid:                           uuid.NewString(),
			Name:                           "",
			IncludeConversationPreferences: true,
		}
		reqByte, err := Json.Marshal(rq)
		if err != nil {
			return nil, err, ""
		}
		resp, err := c.client.R().
			SetHeader("accept", contentTypeAll).
			SetBodyBytes(reqByte).
			Post(createConversationUrl)
		defer resp.Body.Close()
		conversation := &ClaudeConversation{}
		err = Json.NewDecoder(resp.Body).Decode(&conversation)
		if err != nil {
			return nil, err, ""
		}
		conversationId = conversation.ID
	}
	goUrl := "https://claude.ai/api/organizations/" + organizationId + "/chat_conversations/" + conversationId + "/completion"
	resp, err := c.client.R().
		SetHeader("accept", acceptStream).
		SetBodyBytes(ccrByte).
		Post(goUrl)
	return resp, err, conversationId
}

func (c *claude) goChatWeb(ccr *ClaudeChatRequest) (*Stream, error) {
	resp, err, _ := c.doChatWebRequest(ccr)
	if err != nil {
		Log.Error("request web chat error", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logHttpStatusErr(resp, "claude web:request web chat http status error")
		return nil, errors.New("request web chat http status error:" + resp.Status)
	}

	return originStream(resp)
}

// 返回stream流
func (c *claude) goChatWebToApi(ccr *ClaudeChatRequest) (*Stream, error) {
	resp, err, conversationId := c.doChatWebRequest(ccr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logHttpStatusErr(resp, "claude web to api:request web chat http status error")
		return nil, errors.New("request web chat http status error:" + resp.Status)
	}

	// 处理stream流
	parentId := ""
	stream := &Stream{
		Events: make(chan *EventData),
		Closed: make(chan struct{}),
	}
	go func() {
		var ir io.Reader
		if resp.Header.Get("Content-Encoding") == "br" {
			ir = brotli.NewReader(resp.Body)
		} else {
			ir = resp.Body
		}
		reader := bufio.NewReader(ir)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if strings.HasPrefix(line, "data:") {
				raw := line[6:]
				raw = strings.TrimSuffix(raw, "\n")
				var chatRes ClaudeChatResponse
				err = Json.UnmarshalFromString(raw, &chatRes)
				if err != nil {
					continue
				}
				switch chatRes.Type {
				case "message_start":
					parentId = chatRes.Message.Uuid
					stream.Events <- &EventData{Name: "", Data: line}
				case "content_block_delta":
					acsr := &AnthropicChatStreamResponse{
						Type:  "content_block_delta",
						Index: chatRes.Index,
						Delta: &ClaudeChatDelta{
							Type: "text_delta",
							Text: chatRes.Delta.Text,
						},
						Claude: &ClaudeExtResponse{
							ParentId:       parentId,
							ConversationId: conversationId,
						},
					}
					outJson, _ := Json.MarshalToString(acsr)
					stream.Events <- &EventData{Name: "", Data: "data: " + outJson + "\n\n"}
				default:
					stream.Events <- &EventData{Name: "", Data: line}
				}
			}
		}
	}()
	return stream, nil
}

func (c *claude) goChatWebToOpenaiApi(ccr *ClaudeChatRequest) (*Stream, error) {
	resp, err, conversationId := c.doChatWebRequest(ccr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 处理返回
	parentId, resMsgId, model := "", "", ""
	stream := &Stream{
		Events: make(chan *EventData),
		Closed: make(chan struct{}),
	}
	go func() {
		var ir io.Reader
		if resp.Header.Get("Content-Encoding") == "br" {
			ir = brotli.NewReader(resp.Body)
		} else {
			ir = resp.Body
		}
		reader := bufio.NewReader(ir)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if strings.HasPrefix(line, "data:") {
				raw := line[6:]
				raw = strings.TrimSuffix(raw, "\n")
				var chatRes ClaudeChatResponse
				err = Json.UnmarshalFromString(raw, &chatRes)
				if err != nil {
					continue
				}
				switch chatRes.Type {
				case "message_start":
					parentId = chatRes.Message.Uuid
					resMsgId = chatRes.Message.ID
					model = chatRes.Message.Model
				case "content_block_delta":
					var choices []*ChatCompletionChoice
					choices = append(choices, &ChatCompletionChoice{
						Delta: &ChatCompletionMessage{
							Role:    "assistant",
							Content: chatRes.Delta.Text,
						},
						Index: 0,
					})
					outRes := &ChatCompletionResponse{
						ID:      resMsgId,
						Choices: choices,
						Created: time.Now().Unix(),
						Model:   model,
						Object:  "chat.completion.chunk",
						Claude: &ClaudeExtResponse{
							ParentId:       parentId,
							ConversationId: conversationId,
						},
					}
					outJson, _ := Json.MarshalToString(outRes)
					stream.Events <- &EventData{Name: "", Data: "data: " + outJson + "\n\n"}
				}
			}
		}
		stream.Events <- &EventData{Name: "", Data: "data: [DONE]\n\n"}
	}()
	return stream, nil
}

// 通用请求体转成web chat请求体
func (c *claude) commonToClaudeChatRequest(rc *ChatCompletionRequest) (*ClaudeChatRequest, error) {
	if rc.Source == "" && rc.Openai == nil {
		return nil, errors.New("params error")
	}
	if rc.Openai.ClaudeExt == nil {
		rc.Openai.ClaudeExt = &ClaudeExtRequest{}
	}
	prompt := rc.Openai.ParsePromptText()
	if prompt == "" {
		return nil, errors.New("request error:message empty")
	}
	parentMsgId := rc.Openai.ClaudeExt.ParentMessageId
	if parentMsgId == "" {
		parentMsgId = "00000000-0000-4000-8000-000000000000"
	}
	ccr := &ClaudeChatRequest{
		Prompt:             prompt,
		ParentMessageUuid:  parentMsgId,
		Timezone:           "Asia/Shanghai",
		PersonalizedStyles: c.generateClaudePersonalizedStyle(),
		RenderingMode:      "messages",
		ClaudeExt: &ClaudeExtRequest{
			ConversationId: rc.Openai.ClaudeExt.ConversationId,
			OrganizationId: rc.Openai.ClaudeExt.OrganizationId},
	}
	return ccr, nil
}

func (c *claude) generateClaudePersonalizedStyle() []*ClaudePersonalizedStyle {
	var cps []*ClaudePersonalizedStyle
	cps = append(cps, &ClaudePersonalizedStyle{
		Name:      "Normal",
		Prompt:    "Normal",
		Summary:   "Default responses from Claude",
		IsDefault: true,
		Type:      "default",
		Key:       "Default",
	})
	return cps
}

func (c *claude) generateOrganizationId(forceRefresh bool) (string, error) {
	if !forceRefresh && c.config.Auth.organizationId != "" {
		return c.config.Auth.organizationId, nil
	}
	resp, err := c.client.R().
		SetHeader("accept", contentTypeAll).
		Get("https://claude.ai/api/organizations")
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New("request organization http status error:" + resp.Status)
	}
	var res []*ClaudeOrganization
	err = Json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}
	for k := range res {
		organization := res[k]
		if len(organization.Capabilities) > 0 {
			for y := range organization.Capabilities {
				if organization.Capabilities[y] == "chat" {
					c.config.Auth.organizationId = organization.Uuid
					return organization.Uuid, nil
				}
			}
		}
	}
	return "", nil
}

func logHttpStatusErr(res *mreq.Response, source string) {
	b, err := readAllToString(res.Body)
	if err != nil {
		Log.Error(source,
			zap.Error(err),
			zap.String("status", res.Status))
		return
	}
	Log.Error(source,
		zap.String("status", res.Status),
		zap.String("data", b))
}
