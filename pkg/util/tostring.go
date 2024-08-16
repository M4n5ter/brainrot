package util

import "unsafe"

// []byte -> string : https://github.com/golang/go/issues/43752

// ToMakabakaString 会将 bytes 转换为 string，注意修改 bytes 会使 string "可变"
func ToMakabakaString(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

// https://go.dev/cl/520395,
// https://go.dev/cl/520599,
// https://go.dev/cl/520600,
// go1.22 对 string 强制转换 []byte 做了零拷贝优化，但是这个优化是通过编译器的内联和逃逸分析来实现的，并不是所有的场景都能够优化到零拷贝。
// 注意这种转换实现简单，性能高，但是 string 是 2 个字段的，而 slice 是 3 个字段的。
// 这里 string 转 slice 会导致 cap 是从内存中取出来的未知数据，所以不能对转换后的 slice 做跟容量相关的事情，比如 append。
func ToMakabakaBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
