package aichat

type AiCom interface {
	Proxy(string)
	Chat(string) (*Stream, error)
	ChatEasy(string, ...any) (*Stream, error)
	ChatToApi(string) (*Stream, error)
}
