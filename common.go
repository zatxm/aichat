package aichat

import (
	"bytes"
	"io"
	"math"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

var pool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 10240))
	},
}

// 读取http通信返回结果
func readAllToString(r io.Reader) (string, error) {
	buffer := pool.Get().(*bytes.Buffer)
	buffer.Reset()
	_, err := io.Copy(buffer, r)
	if err != nil {
		pool.Put(buffer)
		return "", err
	}
	pool.Put(buffer)
	temp := buffer.Bytes()
	length := len(temp)
	var body []byte
	if cap(temp) > (length + length/10) {
		body = make([]byte, length)
		copy(body, temp)
	} else {
		body = temp
	}
	return *(*string)(unsafe.Pointer(&body)), nil
}

// 获取当前时间戳,保留三位小数
func nowTimePay() float64 {
	currentTime := float64(time.Now().UnixNano()) / 1e9
	return math.Round(currentTime*1000) / 1000
}

func newChatgptClientContextualInfo() *ChatgptClientContextualInfo {
	return &ChatgptClientContextualInfo{
		IsDarkMode:      false,
		TimeSinceLoaded: randint(22, 33),
		PageHeight:      randint(600, 900),
		PageWidth:       randint(500, 800),
		PixelRatio:      1,
		ScreenHeight:    randint(800, 1200),
		ScreenWidth:     randint(1200, 2000)}
}

func randint(min, max int) int {
	if min > max {
		tmpMax := max
		max = min
		min = tmpMax
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func randomFloat() float64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Float64()
}

func stringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
