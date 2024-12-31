package aichat

type AiCommon interface {
	// 设置代理
	SetProxy(string)
	// 设置鉴权token等
	SetAuth(string)
	// 通用chat请求,stream
	Chat(*ChatCompletionRequest) (*Stream, error)
	// 原始api请求
	ChatApi(*ChatCompletionRequest) (*Stream, error)
	// 通用chat请求转Api返回格式
	ChatToApi(*ChatCompletionRequest) (*Stream, error)
	// api转chat请求再转Api返回格式
	ApiCrossChatToApi(*ChatCompletionRequest) (*Stream, error)
	// 通用chat请求(统一请求格式)转成openai api返回格式
	ChatToOpenaiApi(*ChatCompletionRequest) (*Stream, error)
	// api原始json数据字符串请求转Api返回格式
	ApiToOpenaiApi(*ChatCompletionRequest) (*Stream, error)
	// api原始json数据字符串转chat请求再转openai api返回格式
	ApiCrossChatToOpenaiApi(*ChatCompletionRequest) (*Stream, error)
}
