package apperror

// 业务错误码使用五位数字：前三位对应 HTTP 状态码，后两位预留给同类业务错误细分。
// 响应层负责将这些业务错误码映射为实际的 HTTP 状态码。
const (
	CodeOK                 = 0     // 请求处理成功。
	CodeInvalidArgument    = 40000 // 请求参数缺失、格式错误或未通过校验。
	CodeUnauthorized       = 40100 // 请求未提供有效身份凭证。
	CodeForbidden          = 40300 // 身份有效，但没有执行当前操作的权限。
	CodeNotFound           = 40400 // 请求的接口或业务数据不存在。
	CodeMethodNotAllowed   = 40500 // 接口存在，但不支持当前 HTTP 方法。
	CodeConflict           = 40900 // 请求与当前数据状态冲突。
	CodeInternal           = 50000 // 服务端发生未预期的内部错误。
	CodeServiceUnavailable = 50300 // 依赖或服务暂时不可用。
)
