package zjson

import (
	"reflect"
	"unsafe"
)

// string转bytes, 转换后的bytes禁止写, 否则产生运行故障
func string2Bytes(s *string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// bytes转string
func bytes2String(b []byte) *string {
	return (*string)(unsafe.Pointer(&b))
}
