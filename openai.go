package aichat

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/brotli"
	"github.com/google/uuid"
	mreq "github.com/imroc/req/v3"
	"golang.org/x/crypto/sha3"
)

type openai struct {
	config    *Config
	uuid      string
	startTime time.Time
	client    *mreq.Client
}

func NewOpenai(cfg *Config) AiCommon {
	client := mreq.C().SetUserAgent(userAgent).ImpersonateChrome()
	if cfg.ProxyUrl != "" {
		client.SetProxyURL(cfg.ProxyUrl)
	}
	if cfg.Auth == nil {
		cfg.Auth = &Auth{}
	}
	return &openai{config: cfg, client: client}
}

func (o *openai) SetProxy(proxyUrl string) {
	o.config.ProxyUrl = proxyUrl
	if proxyUrl != "" {
		o.client.SetProxyURL(proxyUrl)
	}
}

func (o *openai) SetAuth(auth *Auth) {
	o.config.Auth = auth
}

func (o *openai) Chat(rc *ChatCompletionRequest) (*Stream, error) {
	ccr, err := o.toChatgptCompletionRequest(rc)
	if err != nil {
		return nil, err
	}
	return o.goChatgpt(ccr)
}

func (o *openai) ChatApi(rc *ChatCompletionRequest) (*Stream, error) {
	occr, err := o.toOpenaiChatCompletionRequest(rc)
	rb, err := Json.Marshal(occr)
	if err != nil {
		return nil, err
	}

	// 请求通信
	res, err := o.client.R().
		SetHeader("content-type", contentTypeJson).
		SetHeader("authorization", "Bearer "+o.config.Auth.Token).
		SetBodyBytes(rb).Post(openaiApiUrl + "/chat/completions")
	defer res.Body.Close()

	// 处理错误、或非流
	if res.StatusCode >= 400 || !occr.Stream {
		b, err := readAllToString(res.Body)
		if err != nil {
			return nil, err
		}
		return &Stream{Data: b}, nil
	}

	return originStream(res)
}

func (o *openai) ChatToApi(rc *ChatCompletionRequest) (*Stream, error) {
	ccr, err := o.toChatgptCompletionRequest(rc)
	if err != nil {
		return nil, err
	}
	return o.goChatgptToApi(ccr)
}

func (o *openai) ApiCrossChatToApi(rc *ChatCompletionRequest) (*Stream, error) {
	ccr, err := o.apiToChatgptCompletionRequest(rc)
	if err != nil {
		return nil, err
	}
	return o.goChatgptToApi(ccr)
}

func (o *openai) ChatToOpenaiApi(rc *ChatCompletionRequest) (*Stream, error) {
	return o.ChatToApi(rc)
}

func (o *openai) ApiToOpenaiApi(rc *ChatCompletionRequest) (*Stream, error) {
	return o.ChatApi(rc)
}

func (o *openai) ApiCrossChatToOpenaiApi(rc *ChatCompletionRequest) (*Stream, error) {
	return o.ApiCrossChatToApi(rc)
}

func (o *openai) CommonChatToOpenaiApi(rc *ChatCompletionRequest) (*Stream, error) {
	return o.ApiCrossChatToApi(rc)
}

func (o *openai) toChatgptCompletionRequest(rc *ChatCompletionRequest) (*ChatgptCompletionRequest, error) {
	if rc.Source == "" && rc.Chatgpt == nil {
		return nil, errors.New("params error")
	}
	ccr := &ChatgptCompletionRequest{}
	if rc.Source != "" {
		err := Json.UnmarshalFromString(rc.Source, &ccr)
		if err != nil {
			return nil, err
		}
	} else {
		ccr = rc.Chatgpt
	}
	return ccr, nil
}

func (o *openai) toOpenaiChatCompletionRequest(rc *ChatCompletionRequest) (*OpenaiChatCompletionRequest, error) {
	if rc.Source == "" && rc.Openai == nil {
		return nil, errors.New("params error")
	}
	occr := &OpenaiChatCompletionRequest{}
	if rc.Source != "" {
		err := Json.UnmarshalFromString(rc.Source, &occr)
		if err != nil {
			return nil, err
		}
	} else {
		occr = rc.Openai
	}
	return occr, nil
}

// api通用请求转成chatgpt请求
func (o *openai) apiToChatgptCompletionRequest(rc *ChatCompletionRequest) (*ChatgptCompletionRequest, error) {
	occr, err := o.toOpenaiChatCompletionRequest(rc)
	if err != nil {
		return nil, err
	}

	msg := occr.ParsePromptText()
	if msg == "" {
		return nil, errors.New("request error:message empty")
	}
	// 通信请求体
	if occr.ChatgptExt == nil {
		occr.ChatgptExt = &ChatgptRequest{}
	}
	msgId := occr.ChatgptExt.MessageId //请求信息ID
	if msgId == "" {
		msgId = uuid.NewString()
	}
	parentMsgId := occr.ChatgptExt.ParentMessageId //请求父信息ID
	if parentMsgId == "" {
		parentMsgId = uuid.NewString()
	}
	model := occr.Model
	if model == "" {
		model = "auto"
	}
	var ccrMsg []*ChatgptRequestMessage
	ccrMsg = append(ccrMsg, &ChatgptRequestMessage{
		ID:     msgId,
		Author: &ChatgptAuthor{Role: "user"},
		Content: &ChatgptContent{
			ContentType: "text",
			Parts:       []any{msg}},
		CreateTime: nowTimePay(),
		Metadata:   map[string]any{"custom_symbol_offsets": []any{}},
	})
	ccr := &ChatgptCompletionRequest{
		Action:                           "next",
		ClientContextualInfo:             newChatgptClientContextualInfo(),
		ConversationMode:                 map[string]string{"kind": "primary_assistant"},
		ConversationOrigin:               nil,
		ForceParagen:                     false,
		ForceParagenModelSlug:            "",
		ForceRateLimit:                   false,
		HistoryAndTrainingDisabled:       false,
		Messages:                         ccrMsg,
		Model:                            model,
		ParagenCotSummaryDisplayOverride: "allow",
		ParagenStreamTypeOverride:        nil,
		ParentMessageId:                  parentMsgId,
		ResetRateLimits:                  false,
		Suggestions:                      []string{},
		SupportedEncodings:               []string{"v1"},
		SupportsBuffering:                true,
		SystemHints:                      []any{},
		// Timezone:                         "Asia/Shanghai",
		TimezoneOffsetMin:  -480,
		WebsocketRequestId: uuid.NewString()}
	if occr.ChatgptExt.ConversationId != "" {
		ccr.ConversationId = occr.ChatgptExt.ConversationId //会话ID
	}
	return ccr, nil
}

func (o *openai) goChatRequirement(requireProof string) (*ChatgptRequirement, *mreq.Request, error) {
	if o.uuid == "" {
		o.uuid = uuid.NewString()
		o.client.SetCommonCookies(&http.Cookie{Name: "oai-did", Value: o.uuid, Path: "/", Domain: hostURL.Host})
	}
	rq := o.client.R().
		SetHeader("accept", "*/*").
		SetHeader("content-type", contentTypeJson).
		SetHeader("Oai-Device-Id", o.uuid).
		SetHeader("Oai-Language", "en-US")
	var chatRequirementUrl string
	if o.config.Auth.Token == "" {
		chatRequirementUrl = "https://chatgpt.com/backend-anon/sentinel/chat-requirements"
	} else {
		chatRequirementUrl = "https://chatgpt.com/backend-api/sentinel/chat-requirements"
		rq.SetBearerAuthToken(o.config.Auth.Token)
	}
	jsonBody, _ := Json.Marshal(map[string]string{"p": requireProof})
	resp, err := rq.SetBodyBytes(jsonBody).Post(chatRequirementUrl)
	if err != nil {
		return nil, rq, err
	}
	defer resp.Body.Close()
	var require ChatgptRequirement
	err = Json.NewDecoder(resp.Body).Decode(&require)
	if err != nil {
		return nil, rq, err
	}
	if o.config.Auth.Token == "" && require.ForceLogin {
		return nil, rq, errors.New("Must login")
	}
	return &require, rq, nil
}

func (o *openai) doChatgptRequest(ccr *ChatgptCompletionRequest, rq *mreq.Request, require *ChatgptRequirement) (*mreq.Response, error) {
	reqByte, err := Json.Marshal(ccr)
	if err != nil {
		return nil, err
	}
	/*****通信对话*****/
	var chatUrl string
	if o.config.Auth.Token == "" {
		chatUrl = "https://chatgpt.com/backend-anon/conversation"
	} else {
		chatUrl = "https://chatgpt.com/backend-api/conversation"
	}
	rq.SetHeader("accept", "text/event-stream").
		SetHeader("openai-sentinel-chat-requirements-token", require.Token)
	if require.Proofofwork.Required {
		proofToken := o.parseProofToken(require)
		rq.SetHeader("openai-sentinel-proof-token", proofToken)
	}
	if require.Turnstile.Required {
		turnstileToken := "" //暂时为空，不需要
		rq.SetHeader("openai-sentinel-turnstile-token", turnstileToken)
	}
	return rq.SetHeader("origin", "https://chatgpt.com").
		SetHeader("referer", "https://chatgpt.com/").
		SetBodyBytes(reqByte).
		Post(chatUrl)
}

func (o *openai) goChatgpt(ccr *ChatgptCompletionRequest) (*Stream, error) {
	o.startTime = time.Now()
	requireProof := "gAAAAAC" + o.generateAnswer(strconv.FormatFloat(rand.Float64(), 'f', -1, 64), "0")
	require, rq, err := o.goChatRequirement(requireProof)

	// 通信对话
	resp, err := o.doChatgptRequest(ccr, rq, require)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("request chat return http status error")
	}

	return originStream(resp)
}

// chatgpt转api返回
func (o *openai) goChatgptToApi(ccr *ChatgptCompletionRequest) (*Stream, error) {
	o.startTime = time.Now()
	requireProof := "gAAAAAC" + o.generateAnswer(strconv.FormatFloat(rand.Float64(), 'f', -1, 64), "0")
	require, rq, err := o.goChatRequirement(requireProof)

	// 通信对话
	resp, err := o.doChatgptRequest(ccr, rq, require)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("request chat return http status error:" + resp.Status)
	}

	// 处理返回
	stream := &Stream{
		Events: make(chan *EventData),
		Closed: make(chan struct{}),
	}
	var resMsgId, resMsgParentId, conversationId, model string
	createdAt := time.Now().Unix()
	go func() {
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				close(stream.Closed)
				return
			}
			if strings.HasPrefix(line, "data:") {
				raw := line[6:]
				if !strings.HasPrefix(raw, "[DONE]") {
					raw = strings.TrimSuffix(raw, "\n")
					chatRes := map[string]any{}
					err := Json.UnmarshalFromString(raw, &chatRes)
					if err != nil {
						continue
					}
					if _, ok := chatRes["v"]; ok {
						switch chatRes["v"].(type) {
						case string:
							var choices []*ChatCompletionChoice
							choices = append(choices, &ChatCompletionChoice{
								Delta: &ChatCompletionMessage{
									Role:    "assistant",
									Content: chatRes["v"].(string),
								},
								Index: 0,
							})
							outRes := &ChatCompletionResponse{
								ID:      resMsgId,
								Choices: choices,
								Created: createdAt,
								Model:   model,
								Object:  "chat.completion.chunk",
								Chatgpt: &ChatgptResponse{
									MessageId:       resMsgId,
									ParentMessageId: resMsgParentId,
									ConversationId:  conversationId,
								},
							}
							outJson, _ := Json.MarshalToString(outRes)
							stream.Events <- &EventData{Name: "", Data: "data: " + outJson + "\n\n"}
						case map[string]any:
							var rMsg ChatgptCompletionResponse
							err := Json.UnmarshalFromString(raw, &rMsg)
							if err != nil {
								continue
							}
							if rMsg.V.Message.Author.Role == "assistant" {
								conversationId = rMsg.V.ConversationId
								resMsgId = rMsg.V.Message.Metadata.ParentId
								resMsgParentId = rMsg.V.Message.ID
								model = rMsg.V.Message.Metadata.ModelSlug
								createdAt = int64(rMsg.V.Message.CreateTime)
							}
						}
					}
				} else {
					stream.Events <- &EventData{Name: "", Data: "data: [DONE]\n\n"}
				}
			}
		}
	}()
	return stream, nil
}

func (o *openai) generateAnswer(seed string, diff string) string {
	o.parseDataBuildId()

	rand.New(rand.NewSource(time.Now().UnixNano()))
	core := cores[rand.Intn(3)]
	rand.New(rand.NewSource(time.Now().UnixNano()))
	screen := screens[rand.Intn(3)] + core

	now := time.Now()
	now = now.In(timeLocation)
	parseTime := now.Format(timeLayout) + " GMT+0800 (中国标准时间)"

	rand.Seed(time.Now().UnixNano())
	navigatorKey := navigatorMap[rand.Intn(len(navigatorMap))]

	rand.Seed(time.Now().UnixNano())
	windowKey := windowMap[rand.Intn(len(windowMap))]

	timeNum := (float64(time.Since(startTime).Nanoseconds()) + rand.Float64()) / 1e6

	config := []any{
		screen,
		parseTime,
		nil,
		randomFloat(),
		userAgent,
		nil,
		dataBuildId,
		"en-US",
		"en-US,es-US,en,es",
		0,
		navigatorKey,
		"location",
		windowKey,
		timeNum,
		uuid.NewString(),
		"",
		8,
		time.Now(),
	}

	diffLen := len(diff)
	hasher := sha3.New512()
	for i := 0; i < 500000; i++ {
		config[3] = i
		config[9] = (i + 2) / 2
		jStr, _ := Json.Marshal(config)
		base := base64.StdEncoding.EncodeToString(jStr)
		hasher.Write([]byte(seed + base))
		hash := hasher.Sum(nil)
		hasher.Reset()
		if hex.EncodeToString(hash[:diffLen])[:diffLen] <= diff {
			return base
		}
	}
	return "wQ8Lk5FbGpA2NcR9dShT6gYjU7VxZ4D" + base64.StdEncoding.EncodeToString([]byte(`"`+seed+`"`))
}

func (o *openai) parseProofToken(r *ChatgptRequirement) string {
	proof := o.generateAnswer(r.Proofofwork.Seed, r.Proofofwork.Difficulty)
	return "gAAAAAB" + proof
}

func (o *openai) parseDataBuildId() {
	resp, err := o.client.R().SetHeader("accept", "*/*").Get("https://chatgpt.com/?oai-dm=1")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	doc.Find("html[data-build]").Each(func(i int, s *goquery.Selection) {
		if id, ok := s.Attr("data-build"); ok {
			dataBuildId = id
		}
	})
}

// 处理原始stream流
func originStream(resp *mreq.Response) (*Stream, error) {
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
				close(stream.Closed)
				return
			}
			if strings.HasPrefix(line, "data:") {
				stream.Events <- &EventData{Name: "", Data: line}
			}
		}
	}()
	return stream, nil
}
