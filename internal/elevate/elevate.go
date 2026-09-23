package elevate

import (
	"fmt"
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const cancelledByUser = 1223
const shownNormal = 1

func IsAdmin() bool {
	shell := windows.NewLazySystemDLL("shell32.dll")
	isAdmin := shell.NewProc("IsUserAnAdmin")
	granted, _, _ := isAdmin.Call()
	return granted != 0
}

func RelaunchElevated() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find running program: %w", err)
	}
	verb, err := windows.UTF16PtrFromString("runas")
	if err != nil {
		return fmt.Errorf("encode elevation request: %w", err)
	}
	target, err := windows.UTF16PtrFromString(exePath)
	if err != nil {
		return fmt.Errorf("encode program path: %w", err)
	}
	params, err := argsPointer()
	if err != nil {
		return err
	}
	shell := windows.NewLazySystemDLL("shell32.dll")
	launcher := shell.NewProc("ShellExecuteW")
	launched, _, _ := launcher.Call(0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(target)),
		params, 0, shownNormal)
	if launched <= 32 {
		if launched == cancelledByUser {
			return fmt.Errorf("the UAC prompt was declined")
		}
		return fmt.Errorf("start elevated copy: code %d", launched)
	}
	return nil
}

func argsPointer() (uintptr, error) {
	if len(os.Args) < 2 {
		return 0, nil
	}
	quoted := make([]string, 0, len(os.Args)-1)
	for _, arg := range os.Args[1:] {
		quoted = append(quoted, `"`+arg+`"`)
	}
	params, err := windows.UTF16PtrFromString(strings.Join(quoted, " "))
	if err != nil {
		return 0, fmt.Errorf("encode program arguments: %w", err)
	}
	return uintptr(unsafe.Pointer(params)), nil
}
