package apperrors

import "errors"

var (
	// ErrUnitMismatch 表示设备单位不一致，不能进入采集流程。
	ErrUnitMismatch = errors.New("unit mismatch")
	// ErrInvalidArgument 表示请求参数非法。
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrNotFound 表示请求目标不存在。
	ErrNotFound = errors.New("not found")
	// ErrInvalidStateTransition 表示状态机迁移非法。
	ErrInvalidStateTransition = errors.New("invalid state transition")
)
