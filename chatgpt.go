package openai

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	mreq "github.com/imroc/req/v3"
	"golang.org/x/crypto/sha3"
)

type chatgpt struct {
	token     string
	startTime time.Time
	client    *mreq.Client
}

func NewGptchat(token, proxyUrl string) OpenAiCom {
	client := mreq.C().SetUserAgent(userAgent).ImpersonateChrome()
	if proxyUrl != "" {
		client.SetProxyURL(proxyUrl)
	}
	return &chatgpt{token: token, client: client}
}

func (g *chatgpt) Proxy(proxyUrl string) {
	g.client.SetProxyURL(proxyUrl)
}

func (g *chatgpt) Chat(msg string, args ...any) (*Stream, error) {
	// 通信请求体
	msgId := "" //请求信息ID
	if len(args) > 1 && args[1].(string) != "" {
		msgId = args[1].(string)
	} else {
		msgId = uuid.NewString()
	}
	parentMsgId := "" //请求父信息ID
	if len(args) > 2 && args[2].(string) != "" {
		parentMsgId = args[2].(string)
	} else {
		parentMsgId = uuid.NewString()
	}
	model := "auto"
	if len(args) > 0 && args[0].(string) != "" {
		model = args[0].(string)
	}
	var reqMsg []*ChatgptRequestMessage
	reqMsg = append(reqMsg, &ChatgptRequestMessage{
		ID:     msgId,
		Author: map[string]string{"role": "user"},
		Content: &ChatgptRequestContent{
			ContentType: "text",
			Parts:       []string{msg}},
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
	if len(args) > 3 && args[3].(string) != "" {
		goReq.ConversationId = args[3].(string) //会话ID
	}
	reqBody, err := Json.MarshalToString(goReq)
	if err != nil {
		return nil, err
	}

	return g.goChat(reqBody)
}

// 原始数据请求，传递原始请求json字符串
func (g *chatgpt) ChatSource(sourceReqStr string) (*Stream, error) {
	var chatReq ChatgptCompletionRequest
	if err := Json.UnmarshalFromString(sourceReqStr, &chatReq); err != nil {
		return nil, errors.New("request params error")
	}
	return g.goChat(sourceReqStr)
}

func (g *chatgpt) goChat(goReqJsonStr string) (*Stream, error) {
	g.startTime = time.Now()
	requireProof := "gAAAAAC" + g.generateAnswer(strconv.FormatFloat(rand.Float64(), 'f', -1, 64), "0")

	deviceId := uuid.NewString()
	g.client.SetCommonCookies(&http.Cookie{Name: "oai-did", Value: deviceId, Path: "/", Domain: hostURL.Host})
	rq := g.client.R().
		SetHeader("accept", "*/*").
		SetHeader("content-type", contentTypeJson).
		SetHeader("Oai-Device-Id", deviceId).
		SetHeader("Oai-Language", "en-US")
	var chatRequirementUrl, chatUrl string
	if g.token == "" {
		chatRequirementUrl = "https://chatgpt.com/backend-anon/sentinel/chat-requirements"
		chatUrl = "https://chatgpt.com/backend-anon/conversation"
	} else {
		chatRequirementUrl = "https://chatgpt.com/backend-api/sentinel/chat-requirements"
		chatUrl = "https://chatgpt.com/backend-api/conversation"
		rq.SetBearerAuthToken(g.token)
	}
	jsonBody, _ := Json.Marshal(map[string]string{"p": requireProof})
	resp, err := rq.SetBodyBytes(jsonBody).Post(chatRequirementUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var require ChatRequirement
	err = Json.NewDecoder(resp.Body).Decode(&require)
	if err != nil {
		return nil, err
	}
	if g.token == "" && require.ForceLogin {
		return nil, errors.New("Must login")
	}

	/*****通信对话*****/
	rq.SetHeader("accept", "text/event-stream").
		SetHeader("openai-sentinel-chat-requirements-token", require.Token)
	if require.Proofofwork.Required {
		proofToken := g.parseProofToken(&require)
		rq.SetHeader("openai-sentinel-proof-token", proofToken)
	}
	if require.Turnstile.Required {
		turnstileToken := "" //暂时为空，不需要
		rq.SetHeader("openai-sentinel-turnstile-token", turnstileToken)
	}
	resp, err = rq.SetHeader("origin", "https://chatgpt.com").
		SetHeader("referer", "https://chatgpt.com/").
		SetBody(goReqJsonStr).
		Post(chatUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("request chat return http status error")
	}
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

func (g *chatgpt) parseProofToken(r *ChatRequirement) string {
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
