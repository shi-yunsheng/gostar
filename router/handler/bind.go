package handler

import (
	"encoding/json"
	"errors"
	"reflect"
	"sync"

	"github.com/go-playground/validator/v10"
)

// @en bind parameter
//
// @zh 绑定参数
type Bind struct {
	// @en model, must implement "FromJson(string) error" interface, can implement "Validate() error" interface, if "Validate" interface is implemented, it will be used to validate the model, otherwise use github.com/go-playground/validator/v10 to validate the model, for more information about validator, please refer to https://github.com/go-playground/validator
	//
	// @zh 模型，必须实现"FromJson(string) error"接口，可以选择实现"Validate() error"接口，如果有"Validate"接口，则优先使用"Validate"接口进行校验，否则使用 github.com/go-playground/validator/v10 进行校验，有关validator的用法请参考 https://github.com/go-playground/validator
	Model any
	// @en method, same as HTTP method, if empty, bind according to request method
	//
	// @zh 请求方法，同HTTP方法，为空则根据请求方法进行绑定
	Method string
	// @en ensures model type is initialized only once
	//
	// @zh 确保模型类型只初始化一次
	once sync.Once
	// @en cached model type
	//
	// @zh 缓存模型类型
	modelType reflect.Type
}

// @en validate bind parameter
//
// @zh 验证绑定参数
func (b *Bind) Validate(r *Request) (any, error) {
	// @en initialize model type
	// @zh 初始化模型类型
	b.once.Do(func() {
		t := reflect.TypeOf(b.Model)
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		b.modelType = t
	})

	// @en create model instance
	// @zh 创建模型实例
	modelInstance := reflect.New(b.modelType).Interface()

	// @en if method is empty, bind according to request method
	// @zh 如果方法为空，则根据请求方法进行绑定
	if b.Method == r.Method || b.Method == "" {
		var model map[string]any

		switch r.Method {
		case "POST", "PUT", "PATCH", "DELETE":
			body, err := r.GetAllBody()
			if err != nil {
				return nil, err
			}
			model = body
		default:
			query := r.GetAllQuery()
			model = query
		}

		// @en model must implement "FromJson(string) error" interface
		// @zh 模型必须实现"FromJson(string) error"接口
		if fromJsonObj, ok := modelInstance.(interface{ FromJson(string) error }); ok {
			jsonData, err := json.Marshal(model)
			if err != nil {
				return nil, err
			}
			err = fromJsonObj.FromJson(string(jsonData))
			if err != nil {
				return nil, err
			}

			// @en if model implements "Validate() error" interface, use Validate interface to validate the model
			// @zh 如果模型实现了"Validate() error"接口，则使用Validate接口进行校验
			if validateObj, ok := modelInstance.(interface{ Validate() error }); ok {
				err = validateObj.Validate()
				if err != nil {
					return nil, err
				}
			} else {
				// @en model does not implement "Validate() error" interface, use github.com/go-playground/validator/v10 to validate the model
				// @zh 模型没有实现"Validate() error"接口，使用github.com/go-playground/validator/v10进行校验
				validate := validator.New()
				err = validate.Struct(fromJsonObj)
				if err != nil {
					return nil, err
				}
			}

			modelInstance = fromJsonObj
		} else {
			return nil, errors.New(`model must implement "FromJson(string) error" interface`)
		}
	}
	return modelInstance, nil
}
