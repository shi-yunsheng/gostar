package handler

// @en path parameter
//
// @zh 路径参数
type Param struct {
	// @en parameter key
	//
	// @zh 参数名
	Key string
	// @en parameter value
	//
	// @zh 参数值
	Value any
	// @en parameter default value
	//
	// @zh 参数默认值
	Default any
	// @en parameter type
	//
	// @zh 参数类型
	Type string
	// @en parameter regular expression
	//
	// @zh 参数正则表达式
	Pattern string
	// @en parameter optional
	//
	// @zh 参数可选
	Optional bool
}
