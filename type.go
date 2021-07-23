package zjson

import (
	"errors"
	"fmt"
)

var SyntaxErr = errors.New("语法错误")

type NodeType int

const (
	Null    = NodeType(iota) // 空类型
	Boolean                  // 布尔
	Number                   // 数字
	String                   // 字符串
	Array                    // 数组
	Object                   // 对象
)

func (m NodeType) String() string {
	switch m {
	case Null:
		return "Null"
	case Boolean:
		return "Boolean"
	case Number:
		return "Number"
	case String:
		return "String"
	case Array:
		return "Array"
	case Object:
		return "Object"
	}
	return fmt.Sprintf("undefined<%d>", m)
}
