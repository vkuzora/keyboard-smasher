package main

import (
	"syscall"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procSendInput           = user32.NewProc("SendInput")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procFindWindowW         = user32.NewProc("FindWindowW")
	winmm                   = syscall.NewLazyDLL("winmm.dll")
	procTimeBeginPeriod     = winmm.NewProc("timeBeginPeriod")
	procTimeEndPeriod       = winmm.NewProc("timeEndPeriod")
)

const (
	VK_F1           = 0x70
	INPUT_KEYBOARD  = 1
	KEYEVENTF_KEYUP = 0x0002
)

type KEYBDINPUT struct {
	Vk        uint16
	Scan      uint16
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type INPUT struct {
	Type uint32
	_    uint32
	Ki   KEYBDINPUT
	_    uint64
}

func isKeyPressed(vk int) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return ret&0x8000 != 0
}

func getOurHWND() uintptr {
	title, _ := syscall.UTF16PtrFromString("Keyboard Smasher")
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
	return hwnd
}

func isForeground() bool {
	our := getOurHWND()
	if our == 0 {
		return true
	}
	fg, _, _ := procGetForegroundWindow.Call()
	return fg == our
}
