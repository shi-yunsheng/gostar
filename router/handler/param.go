package handler

// 路径参数
type Param struct {
	// 参数名
	Key string
	// 参数值
	Value any
	// 参数默认值
	Default any
	// 参数类型
	Type string
	// 参数正则表达式
	Pattern string
	// 参数可选
	Optional bool
}
