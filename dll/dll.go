package dll

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

// CallDLLFunc — безопасная универсальная функция вызова любой DLL.
func Dll(dllName, procName string, args ...interface{}) (uintptr, error) {
	dll, err := syscall.LoadDLL(dllName)
	if err != nil {
		return 0, fmt.Errorf("ошибка загрузки DLL '%s': %w", dllName, err)
	}
	defer dll.Release()

	proc, err := dll.FindProc(procName)
	if err != nil {
		return 0, fmt.Errorf("функция '%s' не найдена: %w", procName, err)
	}

	uintptrArgs := make([]uintptr, len(args))
	keepAlive := make([]interface{}, len(args))

	for i, arg := range args {
		ptr, obj := toUintptr(arg)
		uintptrArgs[i] = ptr
		keepAlive[i] = obj
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	r1, _, _ := proc.Call(uintptrArgs...)
	runtime.KeepAlive(keepAlive)

	return r1, nil
}

func toUintptr(arg interface{}) (uintptr, interface{}) {
	switch v := arg.(type) {
	case nil:
		return 0, nil
	case int:
		return uintptr(v), nil
	case uint:
		return uintptr(v), nil
	case uint32:
		return uintptr(v), nil
	case int32:
		return uintptr(v), nil
	case string:
		p, err := syscall.UTF16PtrFromString(v)
		if err != nil {
			return 0, nil
		}
		return uintptr(unsafe.Pointer(p)), p
	case []byte:
		if len(v) == 0 {
			return 0, nil
		}
		return uintptr(unsafe.Pointer(&v[0])), v
	case uintptr:
		return v, nil
	case unsafe.Pointer:
		return uintptr(v), v
	default:
		panic(fmt.Sprintf("неподдерживаемый тип аргумента для DLL: %T", arg))
	}
}