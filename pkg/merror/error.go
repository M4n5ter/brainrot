package merror

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

// DefineError 定义属于当前范围的错误
//
// areaNum: 错误区域编号，moduleNumber: 模块编号，specificNum: 具体错误编号，desensitization: 脱敏后的信息，detail: 详细信息（可选）
func DefineError(areaNum AreaNumber, moduleNumber, specificNum uint32, desensitization string, detail ...string) *Error {
	var code uint32
	switch areaNum {
	case Common:
		code = SpawnCommonError(moduleNumber, specificNum)
	case Client:
		code = SpawnClientError(moduleNumber, specificNum)
	case System:
		code = SpawnSystemError(moduleNumber, specificNum)
	case ThirdParty:
		code = SpawnThirdPartyError(moduleNumber, specificNum)
	default:
		panic("unknown error area")
	}

	RegisterErrorWithDesensitization(code, desensitization)
	return NewError(code, strings.Join(detail, ""))
}

// 未注册的错误码会返回 Unknown error
func Desensitize(code uint32) string {
	if msg, ok := errorMap[code]; ok {
		return msg
	}

	return "Unknown error"
}

func RegisterErrorWithDesensitization(errCode uint32, desensitization string) {
	errorMap[errCode] = desensitization
}

var errorMap = map[uint32]string{}

// 将 error code 映射到发生错误的地方
func MapCodeToModule(code uint32) string {
	areaCode := code / 10000
	moduleCode := (code % 10000) / 100

	var area string
	switch AreaNumber(areaCode * 10000) {
	case Common:
		area = "Common"
	case Client:
		area = "Client"
	case System:
		area = "System"
	case ThirdParty:
		area = "ThirdParty"
	default:
		area = "UnkownArea"
	}

	module := modules[moduleCode]
	return fmt.Sprintf("%s_%s", module, area)
}

// 注册模块编号，成功则原样返回模块编号。模块编号 0<=num<=99，超出范围或者存在重复编号则会 panic
func MustRegisterErrorModule(num ModuleNumber, name string) ModuleNumber {
	if num > 99 {
		log.Fatalln("module number should be less than 100")
	}

	if _, ok := modules[num]; ok {
		log.Fatalf("module %d already exists", num)
	}

	modules[num] = name

	return num
}

var modules = map[ModuleNumber]string{} // thie line equals to var modules = make(map[ModuleNumber]string)

func (e *Error) Wrap(format string, args ...any) error {
	return fmt.Errorf("%s -> %w", fmt.Sprintf(format, args...), e)
}

func NewError(code uint32, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

type Error struct {
	Code    uint32
	Message string
}

func (e *Error) GetCode() uint32 {
	if e != nil {
		return e.Code
	}
	return 0
}

func (e *Error) GetMessage() string {
	if e != nil {
		return e.Message
	}
	return ""
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func UnwrapAll(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}
