package aichat

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var startTime = time.Now()

type FloatMap map[float64]any
type FuncType func(args ...any) any

type orderedMap struct {
	Keys   []string
	Values map[string]any
}

func (o *orderedMap) Add(key string, value any) {
	if _, ok := o.Values[key]; !ok {
		o.Keys = append(o.Keys, key)
	}
	o.Values[key] = value
}

func (o *orderedMap) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	length := len(o.Keys)
	for i := range o.Keys {
		key := o.Keys[i]
		jsonValue, err := Json.Marshal(o.Values[key])
		if err != nil {
			return nil, err
		}
		buffer.WriteString(fmt.Sprintf("\"%s\":%s", key, jsonValue))
		if i < length-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("}")
	return buffer.Bytes(), nil
}

func newOrderedMap() *orderedMap {
	return &orderedMap{
		Values: make(map[string]any),
	}
}

func getTurnstileToken(dx, p string) string {
	b, _ := base64.StdEncoding.DecodeString(dx)
	return processTurnstileToken(bytesToString(b), p)
}

func processTurnstileToken(dx, p string) string {
	result := []rune{}
	pLength := len(p)
	if pLength != 0 {
		for i := range dx {
			result = append(result, rune(int(dx[i])^int(p[i%pLength])))
		}
	} else {
		result = []rune(dx)
	}
	return string(result)
}

func parseTurnstile(dx, p string) string {
	tokens := getTurnstileToken(dx, p)
	var tokenList [][]any
	if err := Json.UnmarshalFromString(tokens, &tokenList); err != nil {
		fmt.Println(err)
		return ""
	}

	var res string
	processMap := getFuncMap()
	processMap[3] = FuncType(func(args ...any) any {
		e := args[0].(string)
		res = base64.StdEncoding.EncodeToString(stringToBytes(e))
		return nil
	})
	processMap[9] = tokenList
	processMap[16] = p
	for _, token := range tokenList {
		e := token[0].(float64)
		t := token[1:]
		f := processMap[e].(FuncType)
		f(t...)
	}

	return res
}

func getFuncMap() FloatMap {
	var processMap FloatMap = FloatMap{}

	processMap[1] = FuncType(func(args ...any) any {
		e := args[0].(float64)
		eStr := toStr(processMap[e])
		tStr := toStr(processMap[args[1].(float64)])
		if eStr != "" && tStr != "" {
			res := processTurnstileToken(eStr, tStr)
			processMap[e] = res
		}
		return nil
	})

	processMap[2] = FuncType(func(args ...any) any {
		e := args[0].(float64)
		processMap[e] = args[1]
		return nil
	})

	processMap[5] = FuncType(func(args ...any) any {
		e := args[0].(float64)
		t := args[1].(float64)
		n := processMap[e]
		tres := processMap[t]
		if n == nil {
			processMap[e] = tres
		} else if isSlice(n) {
			nt := n.([]any)
			nt = append(nt, tres)
			processMap[e] = nt
		} else {
			var res any
			if isString(n) || isString(tres) {
				nstr := toStr(n)
				tstr := toStr(tres)
				res = nstr + tstr
			} else if isFloat64(n) && isFloat64(tres) {
				nnum := n.(float64)
				tnum := tres.(float64)
				res = nnum + tnum
			} else {
				res = "NaN"
			}
			processMap[e] = res
		}
		return nil
	})

	processMap[6] = FuncType(func(args ...any) any {
		e := args[0].(float64)
		t := args[1].(float64)
		n := args[2].(float64)
		tv := processMap[t]
		nv := processMap[n]
		if isString(tv) && isString(nv) {
			res := tv.(string) + "." + nv.(string)
			if res == "window.document.location" {
				processMap[e] = "https://chatgpt.com/"
			} else {
				processMap[e] = res
			}
		}
		return nil
	})

	processMap[7] = FuncType(func(args ...any) any {
		n := []any{}
		l := len(args[1:])
		for i := range l {
			v := args[i+1].(float64)
			n = append(n, processMap[v])
		}
		e := args[0].(float64)
		ev := processMap[e]
		switch ev := ev.(type) {
		case string:
			if ev == "window.Reflect.set" {
				object := n[0].(*orderedMap)
				keyStr := strconv.FormatFloat(n[1].(float64), 'f', -1, 64)
				val := n[2]
				object.Add(keyStr, val)
			}
		case FuncType:
			ev(n...)
		}
		return nil
	})

	processMap[8] = FuncType(func(args ...any) any {
		e := args[0].(float64)
		t := args[1].(float64)
		processMap[e] = processMap[t]
		return nil
	})
	processMap[10] = "window"

	processMap[14] = FuncType(func(args ...any) any {
		e := args[0].(float64)
		t := args[1].(float64)
		tv := processMap[t]
		processMap[e] = nil
		if isString(tv) {
			tokenList := [][]any{}
			if err := Json.UnmarshalFromString(tv.(string), &tokenList); err == nil {
				processMap[e] = tokenList
			}
		}
		return nil
	})

	processMap[15] = FuncType(func(args ...any) any {
		e := args[0].(float64)
		t := args[1].(float64)
		tres, _ := Json.MarshalToString(processMap[t])
		processMap[e] = tres
		return nil
	})

	processMap[17] = FuncType(func(args ...any) any {
		i := []any{}
		l := len(args[2:])
		for k := range l {
			v := args[k+2].(float64)
			i = append(i, processMap[v])
		}
		e := args[0].(float64)
		t := args[1].(float64)
		tv := processMap[t]
		var res any
		switch tv := tv.(type) {
		case string:
			if tv == "window.performance.now" {
				res = (float64(time.Since(startTime).Nanoseconds()) + rand.Float64()) / 1e6
			} else if tv == "window.Object.create" {
				res = newOrderedMap()
			} else if tv == "window.Object.keys" {
				if v, ok := i[0].(string); ok && v == "window.localStorage" {
					res = []string{"STATSIG_LOCAL_STORAGE_INTERNAL_STORE_V4", "STATSIG_LOCAL_STORAGE_STABLE_ID", "client-correlated-secret", "oai/apps/capExpiresAt", "oai-did", "STATSIG_LOCAL_STORAGE_LOGGING_REQUEST", "UiState.isNavigationCollapsed.1"}
				}
			} else if tv == "window.Math.random" {
				rand.NewSource(time.Now().UnixNano())
				res = rand.Float64()
			}
		case FuncType:
			res = tv(i...)
		}
		processMap[e] = res
		return nil
	})

	processMap[18] = FuncType(func(args ...any) any {
		e := args[0].(float64)
		estr := toStr(processMap[e])
		decoded, _ := base64.StdEncoding.DecodeString(estr)
		processMap[e] = bytesToString(decoded)
		return nil
	})

	processMap[19] = FuncType(func(args ...any) any {
		e := args[0].(float64)
		estr := toStr(processMap[e])
		processMap[e] = base64.StdEncoding.EncodeToString(stringToBytes(estr))
		return nil
	})

	processMap[20] = FuncType(func(args ...any) any {
		o := []any{}
		l := len(args[3:])
		for i := range l {
			v := args[i+3].(float64)
			o = append(o, processMap[v])
		}
		e := args[0].(float64)
		t := args[1].(float64)
		n := args[2].(float64)
		ev := processMap[e]
		tv := processMap[t]
		if ev == tv {
			nv := processMap[n]
			switch nv := nv.(type) {
			case FuncType:
				nv(o...)
			}
		}
		return nil
	})

	processMap[21] = FuncType(func(args ...any) any {
		return nil
	})

	processMap[23] = FuncType(func(args ...any) any {
		i := []any{}
		l := len(args[2:])
		for k := range l {
			i = append(i, args[k+2].(float64))
		}
		e := args[0].(float64)
		t := args[1].(float64)
		ev := processMap[e]
		tv := processMap[t]
		if ev != nil {
			switch tv := tv.(type) {
			case FuncType:
				tv(i...)
			}
		}
		return nil
	})

	processMap[24] = FuncType(func(args ...any) any {
		e := args[0].(float64)
		t := args[1].(float64)
		n := args[2].(float64)
		tv := processMap[t]
		nv := processMap[n]
		if isString(tv) && isString(nv) {
			processMap[e] = tv.(string) + "." + nv.(string)
		}
		return nil
	})

	return processMap
}

func isSlice(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
func isFloat64(v any) bool {
	_, ok := v.(float64)
	return ok
}
func isString(v any) bool {
	_, ok := v.(string)
	return ok
}

func toStr(v any) string {
	if v == nil {
		return "undefined"
	} else if isFloat64(v) {
		return strconv.FormatFloat(v.(float64), 'f', -1, 64)
	} else if isString(v) {
		vStr := v.(string)
		if val, ok := specialCases[vStr]; ok {
			return val
		} else {
			return vStr
		}
	} else if slice, ok := v.([]string); ok {
		return strings.Join(slice, ",")
	}
	return ""
}
