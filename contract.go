package openai

type OpenAiCom interface {
	Proxy(string)
	Chat(string, ...any) (*Stream, error)
	ChatSource(string) (*Stream, error)
}
