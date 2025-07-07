package util

import (
	"errors"
	"fmt"
)

// 定义业务错误类型
var (
	ErrInvalidInput      = errors.New("无效的输入参数")
	ErrNotFound          = errors.New("资源不存在")
	ErrInsufficientStock = errors.New("库存不足")
	ErrOrderCreateFailed = errors.New("订单创建失败")
	ErrDatabaseError     = errors.New("数据库操作失败")
	ErrTimeout           = errors.New("操作超时")
)

// BusinessError 业务错误
type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *BusinessError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *BusinessError) Unwrap() error {
	return e.Err
}

// NewBusinessError 创建业务错误
func NewBusinessError(code, message string, err error) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// IsBusinessError 判断是否为业务错误
func IsBusinessError(err error) bool {
	var businessErr *BusinessError
	return errors.As(err, &businessErr)
}

// GetBusinessError 获取业务错误
func GetBusinessError(err error) *BusinessError {
	var businessErr *BusinessError
	if errors.As(err, &businessErr) {
		return businessErr
	}
	return nil
}
