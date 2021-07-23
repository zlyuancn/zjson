package zjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
)

type Node struct {
	Parent      *Node // 父
	FirstChild  *Node // 第一个子
	LastChild   *Node // 最后一个子
	PrevSibling *Node // 上一个同级节点, 如果是Object, 顺序是按照key正序
	NextSibling *Node // 下一个同级节点, 如果是Object, 顺序是按照key正序

	Array  []*Node          // 如果是Array, 这里还会保存所有的节点
	Object map[string]*Node // 如果是Object, 这里还会保存所有的节点

	Type     NodeType    // 节点类型
	Path     string      // 当前节点所在路径
	Key      string      // key, 只有Object存在
	RawValue interface{} // 原始值
	Level    int         // 层级
}

// -------  获取值  ----------

// 获取bool值
func (n *Node) GetBool(def ...bool) bool {
	if n.Type == Boolean {
		return n.RawValue.(bool)
	}
	return len(def) > 0 && def[0]
}

// 获取数值
func (n *Node) GetFloat64(def ...float64) float64 {
	if n.Type == Number {
		return n.RawValue.(float64)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// 获取数值的int值
func (n *Node) GetInt(def ...int) int {
	if n.Type == Number {
		return int(n.RawValue.(float64))
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// 获取字符串值
func (n *Node) GetString(def ...string) string {
	if n.Type == String {
		return n.RawValue.(string)
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// 获取数组
func (n *Node) GetArray(def ...[]interface{}) []interface{} {
	if n.Type == Array {
		return n.RawValue.([]interface{})
	}
	if len(def) > 0 {
		return def[0]
	}
	return nil
}

// 获取切片. 只有非Array类型才会返回默认值. 如果start和end超出范围, 会返回匹配的数据
func (n *Node) GetSlice(start, end int, def ...[]interface{}) []interface{} {
	if n.Type != Array {
		if len(def) > 0 {
			return def[0]
		}
		return nil
	}

	s := n.RawValue.([]interface{})
	if start >= end || start >= len(s) || end <= 0 {
		return make([]interface{}, 0)
	}
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

// 获取Array指定索引的值
func (n *Node) GetIndex(i int, def ...interface{}) interface{} {
	if n.Type == Array {
		v := n.RawValue.([]interface{})
		if i >= 0 && i < len(v) {
			return v[i]
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return nil
}

// 获取Object
func (n *Node) GetObject(def ...map[string]interface{}) map[string]interface{} {
	if n.Type == Object {
		return n.RawValue.(map[string]interface{})
	}
	if len(def) > 0 {
		return def[0]
	}
	return nil
}

// 获取Object指定key的值, 与直接从map中获取值不同的是GetObjectValue在key不存在时也会返回默认值
func (n *Node) GetObjectValue(key string, def ...interface{}) interface{} {
	if n.Type == Object {
		m := n.RawValue.(map[string]interface{})
		if v, ok := m[key]; ok {
			return v
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return nil
}

// 获取数量, 只有Array和Object有效
func (n *Node) GetCount() int {
	if n.Type == Array {
		return len(n.RawValue.([]interface{}))
	}
	if n.Type == Object {
		return len(n.RawValue.(map[string]interface{}))
	}
	return 0
}

// todo 结果查询

// -----  获取节点  --------

// 获取Array的切片节点, 如果start或end超出范围, 会返回匹配的数据. 非Array类型会返回nil
func (n *Node) Slice(start, end int) []*Node {
	if n.Type != Array {
		return nil
	}

	if start >= end || start >= len(n.Array) || end <= 0 {
		return make([]*Node, 0)
	}
	if start < 0 {
		start = 0
	}
	if end > len(n.Array) {
		end = len(n.Array)
	}
	return n.Array[start:end]
}

// 获取Array指定索引的节点, 失败返回nil
func (n *Node) Index(i int) *Node {
	if n.Type == Array {
		if i >= 0 && i < len(n.Array) {
			return n.Array[i]
		}
	}
	return nil
}

// todo 路径查询

// -------  展示  ----------

// 输出
func (n *Node) String() string {
	return n.Type.String() + ":" + n.Path
}

// 转为json文本
func (n *Node) ToJsonText(pretty bool) string {
	switch n.Type {
	case Null:
		return "null"
	case Boolean:
		if n.RawValue.(bool) {
			return "true"
		}
		return "false"
	case Number:
		return fmt.Sprintf("%g", n.RawValue)
	case String:
		return fmt.Sprintf("%q", n.RawValue)
	}

	if pretty {
		out, _ := json.MarshalIndent(&n.RawValue, "", DefaultPrettyFormatIndent)
		return *bytes2String(out)
	}

	out, _ := json.Marshal(&n.RawValue)
	return *bytes2String(out)
}

// -------  加载  ----------

func parseValue(x interface{}, node *Node) {
	// 设置关系
	setRelation := func(n *Node) {
		n.Parent = node
		if node.FirstChild == nil {
			node.FirstChild = n
			node.LastChild = n
		} else {
			t := node.LastChild
			t.NextSibling = n
			n.PrevSibling = t
			node.LastChild = n
		}
	}

	switch v := x.(type) {
	case nil:
		node.Type = Null
	case bool:
		node.Type = Boolean
	case float64:
		node.Type = Number
	case string:
		node.Type = String
	case []interface{}:
		node.Type = Array
		node.Array = make([]*Node, len(v))
		for i, child := range v {
			n := &Node{
				Path:     fmt.Sprintf("%s[%d]", node.Path, i),
				RawValue: child,
				Level:    node.Level + 1,
			}
			setRelation(n)
			parseValue(child, n)
			node.Array[i] = n
		}
	case map[string]interface{}:
		node.Type = Object
		node.Object = make(map[string]*Node, len(v))
		var keys []string
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			child := v[key]
			n := &Node{
				Path:     "/" + key,
				Key:      key,
				RawValue: child,
				Level:    node.Level + 1,
			}
			if n.Level > 1 {
				n.Path = node.Path + n.Path
			}
			setRelation(n)
			parseValue(child, n)
			node.Object[key] = n
		}
	}
}

// 从文件中加载
func LoadFile(filename string) (*Node, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return LoadReader(f)
}

// 从Reader中加载
func LoadReader(r io.Reader) (*Node, error) {
	var x interface{}
	err := json.NewDecoder(r).Decode(&x)
	if err != nil {
		return nil, err
	}
	node := &Node{Path: "/", RawValue: x, Level: 0}
	parseValue(x, node)
	return node, nil
}

// 从字符串中加载
func LoadString(s string) (*Node, error) {
	return LoadReader(bytes.NewReader(string2Bytes(&s)))
}

// 从bytes中加载
func Load(bs []byte) (*Node, error) {
	return LoadReader(bytes.NewReader(bs))
}
