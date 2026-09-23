package banner

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	reset         = "\x1b[0m"
	edgeColor     = "\x1b[38;2;243;239;232m"
	titleColor    = "\x1b[38;2;246;240;228m"
	accentColor   = "\x1b[38;2;153;204;255m"
	fallbackWidth = 80
	clearSeq      = "\x1b[2J\x1b[H"
)

const (
	bannerTitle    = "yourpov.dev"
	bannerSubtitle = "Wi-Fi MAC Changer"
)

// -11 wont fit uintptr which is why im using 10
const stdOutputHandle = ^uintptr(10)

const (
	utf8CodePage                  = 65001
	enableVirtualTerminalHandling = 0x0004
)

func Show() {
	contentWidth, margin := layout()
	bar := strings.Repeat("─", contentWidth)
	divider := strings.Repeat("─", contentWidth-2)

	fmt.Println()
	fmt.Println(margin + edgeColor + "┌" + bar + "┐" + reset)
	fmt.Println(textRow(margin, titleColor+bannerTitle+reset, bannerTitle, contentWidth))
	fmt.Println(margin + edgeColor + "│ " + reset + divider + edgeColor + " │" + reset)
	fmt.Println(textRow(margin, accentColor+bannerSubtitle+reset, bannerSubtitle, contentWidth))
	fmt.Println(margin + edgeColor + "└" + bar + "┘" + reset)
	fmt.Println()
}

func Accent(text string) string {
	return accentColor + text + reset
}

func Clear() {
	enableWindowsConsole()
	fmt.Print(clearSeq)
}

func SetTitle(title string) {
	if runtime.GOOS != "windows" {
		return
	}
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleTitle := kernel32.NewProc("SetConsoleTitleW")
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	setConsoleTitle.Call(uintptr(unsafe.Pointer(titlePtr)))
}

func layout() (contentWidth int, margin string) {
	contentWidth = max(len(bannerTitle), len(bannerSubtitle)) + 4
	margin = strings.Repeat(" ", max((consoleWidth()-contentWidth-2)/2, 0))
	return contentWidth, margin
}

func textRow(margin, rendered, plain string, width int) string {
	left := (width - len(plain)) / 2
	right := width - len(plain) - left
	return margin + edgeColor + "│" + reset +
		strings.Repeat(" ", left) + rendered + strings.Repeat(" ", right) +
		edgeColor + "│" + reset
}

func consoleWidth() int {
	if width, ok := windowsConsoleWidth(); ok {
		return width
	}
	if columns, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && columns > 0 {
		return columns
	}
	return fallbackWidth
}

type consoleWindowRect struct {
	Left, Top, Right, Bottom int16
}

type consoleScreenBufferInfo struct {
	Size              [2]int16
	CursorPosition    [2]int16
	Attributes        uint16
	Window            consoleWindowRect
	MaximumWindowSize [2]int16
}

func windowsConsoleWidth() (int, bool) {
	if runtime.GOOS != "windows" {
		return 0, false
	}
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	getScreenBufferInfo := kernel32.NewProc("GetConsoleScreenBufferInfo")

	stdout, _, _ := getStdHandle.Call(uintptr(stdOutputHandle))
	if stdout == ^uintptr(0) {
		return 0, false
	}
	var info consoleScreenBufferInfo
	ok, _, _ := getScreenBufferInfo.Call(stdout, uintptr(unsafe.Pointer(&info)))
	if ok == 0 {
		return 0, false
	}
	width := int(info.Window.Right-info.Window.Left) + 1
	if width <= 0 {
		return 0, false
	}
	return width, true
}

func enableWindowsConsole() {
	if runtime.GOOS != "windows" {
		return
	}
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	setOutputCP := kernel32.NewProc("SetConsoleOutputCP")

	stdout, _, _ := getStdHandle.Call(uintptr(stdOutputHandle))
	var mode uint32
	ok, _, _ := getConsoleMode.Call(stdout, uintptr(unsafe.Pointer(&mode)))
	if ok != 0 {
		setConsoleMode.Call(stdout, uintptr(mode|enableVirtualTerminalHandling))
	}
	setOutputCP.Call(uintptr(utf8CodePage))
}
