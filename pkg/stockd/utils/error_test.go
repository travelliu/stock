package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"stock/pkg/stockd/utils"
)

func TestNew(t *testing.T) {
	err := utils.New(utils.ErrInvalidParam, "bad value %v", 42)
	assert.Equal(t, utils.ErrInvalidParam, err.(interface{ Code() int }).Code())
	assert.Contains(t, err.Error(), "bad value 42")
}

func TestWrap(t *testing.T) {
	inner := utils.New(utils.ErrUserNotFound, "user missing")
	wrapped := utils.Wrap(utils.ErrInvalidParam, inner, "outer")
	assert.Contains(t, wrapped.Error(), "outer")
	assert.Contains(t, wrapped.Error(), "user missing")
}

func TestGetCodeAndData(t *testing.T) {
	err := utils.New(utils.ErrInvalidParam, "msg", 1, 2)
	code, data := utils.GetCodeAndData(err)
	assert.Equal(t, utils.ErrInvalidParam, code)
	assert.Equal(t, []interface{}{1, 2}, data)
}

func TestGetCodeAndData_StandardError(t *testing.T) {
	code, data := utils.GetCodeAndData(assert.AnError)
	assert.Equal(t, 0, code)
	assert.Nil(t, data)
}
