package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator 封装了validator实例
type Validator struct {
	validator *validator.Validate
}

// NewValidator 创建新的验证器实例
func NewValidator() *Validator {
	v := validator.New()

	// 注册字段名称函数
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validator: v}
}

// Validate 验证结构体
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

// ValidateVar 验证单个变量
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validator.Var(field, tag)
}

// GetValidationErrors 获取详细的验证错误信息
func (v *Validator) GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			param := e.Param()

			// 生成用户友好的错误消息
			message := getErrorMessage(field, tag, param)
			errors[field] = message
		}
	}

	return errors
}

// getErrorMessage 生成用户友好的错误消息
func getErrorMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// 全局验证器实例
var DefaultValidator = NewValidator()

// 便捷函数

// Validate 验证结构体
func Validate(i interface{}) error {
	return DefaultValidator.Validate(i)
}

// ValidateVar 验证单个变量
func ValidateVar(field interface{}, tag string) error {
	return DefaultValidator.ValidateVar(field, tag)
}

// GetValidationErrors 获取详细的验证错误信息
func GetValidationErrors(err error) map[string]string {
	return DefaultValidator.GetValidationErrors(err)
}
