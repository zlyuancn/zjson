package zjson

import (
	"bytes"
	"encoding/json"
)

// 默认换行间隔
var DefaultPrettyFormatIndent = "   "

// 将json文本转换为好看一点的输出
func JsonPretty(s, indent string) (string, error) {
	return JsonPrettyBytes(string2Bytes(&s), indent)
}

// 将json字节数组转换为好看一点的输出
func JsonPrettyBytes(bs []byte, indent string) (string, error) {
	if indent == "" {
		indent = DefaultPrettyFormatIndent
	}

	var out bytes.Buffer
	err := json.Indent(&out, bs, "", indent)
	return out.String(), err
}

// 将对象格式化为json字符串
func JsonFormatObj(v interface{}, indent string) (string, error) {
	bs, err := json.MarshalIndent(v, "", indent)
	return *bytes2String(bs), err
}
