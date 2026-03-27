package response

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"strings"
)

const (
	SUCCESS = 0
	FAILED  = -1
	ERROR   = -2
)

func ResultErrors(c *fiber.Ctx, status, code int, msg map[string]string) error {
	return c.Status(status).JSON(fiber.Map{
		"code": code,
		"msg":  msg,
	})
}

func ResultError(c *fiber.Ctx, status, code int, msg string) error {
	return c.Status(status).JSON(fiber.Map{
		"code": code,
		"msg":  msg,
	})
}

// resultOkSingle 单对象响应，不含 count 字段
func resultOkSingle(c *fiber.Ctx, code int, msg string, data interface{}) error {
	return c.JSON(fiber.Map{
		"code": code,
		"msg":  msg,
		"data": data,
	})
}

// ResultOk 列表响应，含 count 字段（count 为数据库总记录数，用于分页）
func ResultOk(c *fiber.Ctx, code int, msg string, count int64, data interface{}) error {
	return c.JSON(fiber.Map{
		"code":  code,
		"msg":   msg,
		"count": count,
		"data":  data,
	})
}

// Ok returns 200 ok
func Ok(c *fiber.Ctx) error {
	return resultOkSingle(c, SUCCESS, "success", map[string]interface{}{})
}

// OkWithMessage returns 200 ok and message in content
func OkWithMessage(c *fiber.Ctx, message string) error {
	return resultOkSingle(c, SUCCESS, message, map[string]interface{}{})
}

// OkWithData returns 200 ok and single object，不含 count
func OkWithData(c *fiber.Ctx, data interface{}) error {
	return resultOkSingle(c, SUCCESS, "success", data)
}

// OkWithMsgData returns 200 ok and single object with message，不含 count
func OkWithMsgData(c *fiber.Ctx, message string, data interface{}) error {
	return resultOkSingle(c, SUCCESS, message, data)
}

// OkWithList returns 200 ok and paginated list，count 为数据库总记录数
func OkWithList(c *fiber.Ctx, count int64, data interface{}) error {
	return ResultOk(c, SUCCESS, "success", count, data)
}

// FailWithMsg returns 200 ok and failed message in content
func FailWithMsg(c *fiber.Ctx, message string) error {
	return resultOkSingle(c, FAILED, message, map[string]interface{}{})
}

// FailWithError returns 200 ok and failed message in content
func FailWithError(c *fiber.Ctx, message string, err error) error {
	var builder strings.Builder
	builder.WriteString(message)
	builder.WriteString(": ")
	builder.WriteString(err.Error())
	return resultOkSingle(c, FAILED, builder.String(), nil)
}

// FailWithStatus return status and status message in content
func FailWithStatus(c *fiber.Ctx, status int) error {
	return ResultError(c, status, FAILED, utils.StatusMessage(status))
}

// BadRequest Return status 400 and single error message in content
func BadRequest(c *fiber.Ctx, message string) error {
	return ResultError(c, fiber.StatusBadRequest, FAILED, message)
}

// BadRequests Return status 400 and multi error message in content
func BadRequests(c *fiber.Ctx, msg map[string]string) error {
	return ResultErrors(c, fiber.StatusBadRequest, FAILED, msg)
}

// ServerError returns 500 and error message,err can be nil
func ServerError(c *fiber.Ctx, message string, err error) error {
	if err == nil {
		return ResultError(c, fiber.StatusInternalServerError, ERROR, message)
	}
	var builder strings.Builder
	builder.WriteString(message)
	builder.WriteString(": ")
	builder.WriteString(err.Error())
	return ResultError(c, fiber.StatusInternalServerError, ERROR, builder.String())
}

// ServerErrors returns 500 and multi error message
func ServerErrors(c *fiber.Ctx, message map[string]string) error {
	return ResultErrors(c, fiber.StatusInternalServerError, ERROR, message)
}

// UnAuthorizedError returns 401 and error message
func UnAuthorizedError(c *fiber.Ctx, message string) error {
	return ResultError(c, fiber.StatusUnauthorized, ERROR, message)
}

// UnAuthorizedErrors returns 401 and multi error message
func UnAuthorizedErrors(c *fiber.Ctx, message map[string]string) error {
	return ResultErrors(c, fiber.StatusUnauthorized, ERROR, message)
}
