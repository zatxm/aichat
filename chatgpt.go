package aichat

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	mreq "github.com/imroc/req/v3"
	"golang.org/x/crypto/sha3"
)

type chatgpt struct {
	token     string
	proxyUrl  string
	uuid      string
	startTime time.Time
	client    *mreq.Client
}

func NewChatgpt(token, proxyUrl string) AiCommon {
	client := mreq.C().SetUserAgent(userAgent).ImpersonateChrome()
	if proxyUrl != "" {
		client.SetProxyURL(proxyUrl)
	}
	return &chatgpt{token: token, proxyUrl: proxyUrl, client: client}
}

func (g *chatgpt) SetProxy(proxyUrl string) {
	g.proxyUrl = proxyUrl
	if g.proxyUrl != "" {
		g.client.SetProxyURL(proxyUrl)
	}
}

func (g *chatgpt) SetAuth(token string) {
	g.token = token
}

func (g *chatgpt) Chat(rc *ChatCompletionRequest) (*Stream, error) {
	var goReqStr string
	if rc.Source != "" {
		goReq := &ChatgptCompletionRequest{}
		err := Json.UnmarshalFromString(rc.Source, &goReq)
		if err != nil {
			return nil, err
		}
		goReqStr = rc.Source
	} else {
		msg := rc.ParsePromptText()
		if msg == "" {
			return nil, errors.New("request error:message empty")
		}
		// 通信请求体
		if rc.Chatgpt == nil {
			rc.Chatgpt = &ChatgptRequest{}
		}
		msgId := rc.Chatgpt.MessageId //请求信息ID
		if msgId == "" {
			msgId = uuid.NewString()
		}
		parentMsgId := rc.Chatgpt.ParentMessageId //请求父信息ID
		if parentMsgId == "" {
			parentMsgId = uuid.NewString()
		}
		model := rc.Model
		if model == "" {
			model = "auto"
		}
		var reqMsg []*ChatgptRequestMessage
		reqMsg = append(reqMsg, &ChatgptRequestMessage{
			ID:     msgId,
			Author: &ChatgptAuthor{Role: "user"},
			Content: &ChatgptContent{
				ContentType: "text",
				Parts:       []any{msg}},
			CreateTime: nowTimePay(),
			Metadata:   map[string]any{"custom_symbol_offsets": []any{}},
		})
		goReq := &ChatgptCompletionRequest{
			Action:                           "next",
			ClientContextualInfo:             newChatgptClientContextualInfo(),
			ConversationMode:                 map[string]string{"kind": "primary_assistant"},
			ConversationOrigin:               nil,
			ForceParagen:                     false,
			ForceParagenModelSlug:            "",
			ForceRateLimit:                   false,
			HistoryAndTrainingDisabled:       false,
			Messages:                         reqMsg,
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
		if rc.Chatgpt.ConversationId != "" {
			goReq.ConversationId = rc.Chatgpt.ConversationId //会话ID
		}
		goReqStr, _ = Json.MarshalToString(goReq)
	}

	return g.goChatgpt(goReqStr)
}

func (g *chatgpt) ChatApi(rc *ChatCompletionRequest) (*Stream, error) {
	if rc.Source != "" {
		err := Json.UnmarshalFromString(rc.Source, &rc)
		if err != nil {
			return nil, err
		}
	}
	rc.Stream = true
	rb, err := Json.Marshal(rc)
	if err != nil {
		return nil, err
	}

	// 请求通信
	res, err := g.client.R().
		SetHeader("content-type", contentTypeJson).
		SetHeader("authorization", "Bearer "+g.token).
		SetBodyBytes(rb).Post(openaiApiUrl + "/chat/completions")
	defer res.Body.Close()

	// 处理错误
	if res.StatusCode >= 400 {
		b, err := readAllToString(res.Body)
		if err != nil {
			return nil, err
		}
		return &Stream{Data: b}, nil
	}

	// 处理返回数据
	stream := &Stream{
		Events: make(chan *EventData),
		Closed: make(chan struct{}),
	}
	go func() {
		reader := bufio.NewReader(res.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if line == "\n" {
				continue
			}
			stream.Events <- &EventData{Name: "", Data: line}
		}
	}()
	return stream, nil
}

func (g *chatgpt) ChatToApi(rc *ChatCompletionRequest) (*Stream, error) {
	var goReqStr string
	if rc.Source != "" {
		goReq := &ChatgptCompletionRequest{}
		err := Json.UnmarshalFromString(rc.Source, &goReq)
		if err != nil {
			return nil, err
		}
		goReqStr = rc.Source
	} else {
		msg := rc.ParsePromptText()
		if msg == "" {
			return nil, errors.New("request error:message empty")
		}
		// 通信请求体
		if rc.Chatgpt == nil {
			rc.Chatgpt = &ChatgptRequest{}
		}
		msgId := rc.Chatgpt.MessageId //请求信息ID
		if msgId == "" {
			msgId = uuid.NewString()
		}
		parentMsgId := rc.Chatgpt.ParentMessageId //请求父信息ID
		if parentMsgId == "" {
			parentMsgId = uuid.NewString()
		}
		model := rc.Model
		if model == "" {
			model = "auto"
		}
		var reqMsg []*ChatgptRequestMessage
		reqMsg = append(reqMsg, &ChatgptRequestMessage{
			ID:     msgId,
			Author: &ChatgptAuthor{Role: "user"},
			Content: &ChatgptContent{
				ContentType: "text",
				Parts:       []any{msg}},
			CreateTime: nowTimePay(),
			Metadata:   map[string]any{"custom_symbol_offsets": []any{}},
		})
		goReq := &ChatgptCompletionRequest{
			Action:                           "next",
			ClientContextualInfo:             newChatgptClientContextualInfo(),
			ConversationMode:                 map[string]string{"kind": "primary_assistant"},
			ConversationOrigin:               nil,
			ForceParagen:                     false,
			ForceParagenModelSlug:            "",
			ForceRateLimit:                   false,
			HistoryAndTrainingDisabled:       false,
			Messages:                         reqMsg,
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
		if rc.Chatgpt.ConversationId != "" {
			goReq.ConversationId = rc.Chatgpt.ConversationId //会话ID
		}
		goReqStr, _ = Json.MarshalToString(goReq)
	}

	return g.goChatgptToApi(goReqStr)
}

func (g *chatgpt) ChatToApiSource(sourceReqStr string) (*Stream, error) {
	return nil, nil
}

func (g *chatgpt) ApiCrossChatToApi(apiReqStr string) (*Stream, error) {
	return nil, nil
}

func (g *chatgpt) ChatToOpenai(rc *ChatCompletionRequest) (*Stream, error) {
	return nil, nil
}

func (g *chatgpt) ChatToOpenaiSource(sourceReqStr string) (*Stream, error) {
	return nil, nil
}

func (g *chatgpt) ApiToOpenai(apiReqStr string) (*Stream, error) {
	return nil, nil
}

func (g *chatgpt) ApiCrossChatToOpenai(apiReqStr string) (*Stream, error) {
	return nil, nil
}

func (g *chatgpt) goChatRequirement(requireProof string) (*ChatgptRequirement, *mreq.Request, error) {
	if g.uuid == "" {
		g.uuid = uuid.NewString()
		g.client.SetCommonCookies(&http.Cookie{Name: "oai-did", Value: g.uuid, Path: "/", Domain: hostURL.Host})
	}
	rq := g.client.R().
		SetHeader("accept", "*/*").
		SetHeader("content-type", contentTypeJson).
		SetHeader("Oai-Device-Id", g.uuid).
		SetHeader("Oai-Language", "en-US")
	var chatRequirementUrl string
	if g.token == "" {
		chatRequirementUrl = "https://chatgpt.com/backend-anon/sentinel/chat-requirements"
	} else {
		chatRequirementUrl = "https://chatgpt.com/backend-api/sentinel/chat-requirements"
		rq.SetBearerAuthToken(g.token)
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
	if g.token == "" && require.ForceLogin {
		return nil, rq, errors.New("Must login")
	}
	return &require, rq, nil
}

func (g *chatgpt) doChatgptRequest(reqStr string, rq *mreq.Request, require *ChatgptRequirement) (*mreq.Response, error) {
	/*****通信对话*****/
	var chatUrl string
	if g.token == "" {
		chatUrl = "https://chatgpt.com/backend-anon/conversation"
	} else {
		chatUrl = "https://chatgpt.com/backend-api/conversation"
	}
	rq.SetHeader("accept", "text/event-stream").
		SetHeader("openai-sentinel-chat-requirements-token", require.Token)
	if require.Proofofwork.Required {
		proofToken := g.parseProofToken(require)
		rq.SetHeader("openai-sentinel-proof-token", proofToken)
	}
	if require.Turnstile.Required {
		turnstileToken := "" //暂时为空，不需要
		rq.SetHeader("openai-sentinel-turnstile-token", turnstileToken)
	}
	return rq.SetHeader("origin", "https://chatgpt.com").
		SetHeader("referer", "https://chatgpt.com/").
		SetBody(reqStr).
		Post(chatUrl)
}

func (g *chatgpt) goChatgpt(goReqJsonStr string) (*Stream, error) {
	g.startTime = time.Now()
	requireProof := "gAAAAAC" + g.generateAnswer(strconv.FormatFloat(rand.Float64(), 'f', -1, 64), "0")
	require, rq, err := g.goChatRequirement(requireProof)

	// 通信对话
	resp, err := g.doChatgptRequest(goReqJsonStr, rq, require)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("request chat return http status error")
	}

	// 处理返回
	stream := &Stream{
		Events: make(chan *EventData),
		Closed: make(chan struct{}),
	}
	go func() {
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if line == "\n" {
				continue
			}
			stream.Events <- &EventData{Name: "", Data: line}
		}
	}()
	return stream, nil
}

func (g *chatgpt) goChatgptToApi(goReqJsonStr string) (*Stream, error) {
	g.startTime = time.Now()
	requireProof := "gAAAAAC" + g.generateAnswer(strconv.FormatFloat(rand.Float64(), 'f', -1, 64), "0")
	require, rq, err := g.goChatRequirement(requireProof)

	// 通信对话
	resp, err := g.doChatgptRequest(goReqJsonStr, rq, require)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("request chat return http status error")
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
				break
			}
			if line == "\n" {
				continue
			}
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
						stream.Events <- &EventData{Name: "", Data: outJson + "\n\n"}
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
			}
		}
		stream.Events <- &EventData{Name: "", Data: "data: [DONE]\n\n"}
	}()
	return stream, nil
}

func (g *chatgpt) generateAnswer(seed string, diff string) string {
	g.parseDataBuildId()

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

func (g *chatgpt) parseProofToken(r *ChatgptRequirement) string {
	proof := g.generateAnswer(r.Proofofwork.Seed, r.Proofofwork.Difficulty)
	return "gAAAAAB" + proof
}

func (g *chatgpt) parseDataBuildId() {
	resp, err := g.client.R().SetHeader("accept", "*/*").Get("https://chatgpt.com/?oai-dm=1")
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
