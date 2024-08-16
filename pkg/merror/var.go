// # Error code
//
// We use error code to identify the error type. The error code is a 5-digit integer. The error code is divided into five ranges.
//
// [ common/client/system/third-party/unknown ]x1 [ module ]x2 [ specific error ]x2
//
// For example, 10001 means common error, user module, user exist.
//
// # 0-10000
//
// Reserved.
//
// # 10000-19999
//
// Common error (e.g., database error)
//
// # 20000-29999
//
// Client side error (e.g., invalid input, operation not allowed)
//
// # 30000-39999
//
// System error (e.g., server error)
//
// # 40000-49999
//
// Error from third-party services (e.g., payment error)
package merror

type AreaNumber uint32

const (
	Common     AreaNumber = 10000
	Client     AreaNumber = 20000
	System     AreaNumber = 30000
	ThirdParty AreaNumber = 40000
)

// 0-99
type ModuleNumber = uint32
