package openai

type OpenAiCom interface {
	Proxy(string)
	Chat(string, ...any) (*Stream, error)
}
