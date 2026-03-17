package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procSendInput        = user32.NewProc("SendInput")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")

	winmm               = syscall.NewLazyDLL("winmm.dll")
	procTimeBeginPeriod = winmm.NewProc("timeBeginPeriod")
	procTimeEndPeriod   = winmm.NewProc("timeEndPeriod")
)

const (
	INPUT_KEYBOARD  = 1
	KEYEVENTF_KEYUP = 0x0002
	VK_ESCAPE       = 0x1B
	VK_F1           = 0x70
	VK_F2           = 0x71
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

var allKeys = []byte{
	0x30, //0 key
	0x31, //1 key
	0x32, //2 key
	0x33, //3 key
	0x34, //4 key
	0x35, //5 key
	0x36, //6 key
	0x37, //7 key
	0x38, //8 key
	0x39, //9 key
	0x41, //A key
	0x42, //B key
	0x43, //C key
	0x44, //D key
	0x45, //E key
	0x46, //F key
	0x47, //G key
	0x48, //H key
	0x49, //I key
	0x4A, //J key
	0x4B, //K key
	0x4C, //L key
	0x4D, //M key
	0x4E, //N key
	0x4F, //O key
	0x50, //P key
	0x51, //Q key
	0x52, //R key
	0x53, //S key
	0x54, //T key
	0x55, //U key
	0x56, //V key
	0x57, //W key
	0x58, //X key
	0x59, //Y key
	0x5A, //Z key
	0x60, //Numeric keypad 0 key
	0x61, //Numeric keypad 1 key
	0x62, //Numeric keypad 2 key
	0x63, //Numeric keypad 3 key
	0x64, //Numeric keypad 4 key
	0x65, //Numeric keypad 5 key
	0x66, //Numeric keypad 6 key
	0x67, //Numeric keypad 7 key
	0x68, //Numeric keypad 8 key
	0x69, //Numeric keypad 9 key
	0x70, //F1 key
	0x72, //F3 key
	0x73, //F4 key
	0x74, //F5 key
	0x75, //F6 key
	0x76, //F7 key
	0x77, //F8 key
	0x78, //F9 key
	0x79, //F10 key
	0x7B, //F12 key
	0x7C, //F13 key
	0x7D, //F14 key
	0x7E, //F15 key
	0x7F, //F16 key
	0x80, //F17 key
	0x81, //F18 key
	0x82, //F19 key
	0x83, //F20 key
	0x84, //F21 key
	0x85, //F22 key
	0x86, //F23 key
	0x87, //F24 key
	0x88, //Reserved
	0x89, //Reserved
	0x8A, //Reserved
	0x8B, //Reserved
	0x8C, //Reserved
	0x8D, //Reserved
	0x8E, //Reserved
	0x8F, //Reserved
	0x1C, //IME convert
	0x1D, //IME nonconvert
	0x1E, //IME accept
	0x1F, //IME mode change request
	0x15, //IME Kana mode
	0x17, //IME Junja mode
	0x18, //IME final mode
	0x19, //IME Hanja mode
	0x92, //OEM specific
	0x93, //OEM specific
	0x94, //OEM specific
	0x95, //OEM specific
	0x96, //OEM specific
	0xE3, //OEM specific
	0xE4, //OEM specific
	0xE5, //IME PROCESS key
	0xE6, //OEM specific
	0xE9, //OEM specific
	0xEA, //OEM specific
	0xEB, //OEM specific
	0xEC, //OEM specific
	0xED, //OEM specific
	0xEE, //OEM specific
	0xEF, //OEM specific
	0xF0, //OEM specific
	0xF1, //OEM specific
	0xF2, //OEM specific
	0xF3, //OEM specific
	0xF4, //OEM specific
	0xF5, //OEM specific
	0xE1, //OEM specific
	0xBA, //It can vary by keyboard. For the US ANSI keyboard , the Semiсolon and Colon key
	0xBB, //For any country/region, the Equals and Plus key
	0xBC, //For any country/region, the Comma and Less Than key
	0xBD, //For any country/region, the Dash and Underscore key
	0xBE, //For any country/region, the Period and Greater Than key
	0xBF, //It can vary by keyboard. For the US ANSI keyboard, the Forward Slash and Question Mark key
	0xC0, //It can vary by keyboard. For the US ANSI keyboard, the Grave Accent and Tilde key
	0xDB, //It can vary by keyboard. For the US ANSI keyboard, the Left Brace key
	0xDC, //It can vary by keyboard. For the US ANSI keyboard, the Backslash and Pipe key
	0xDD, //It can vary by keyboard. For the US ANSI keyboard, the Right Brace key
	0xDE, //It can vary by keyboard. For the US ANSI keyboard, the Apostrophe and Double Quotation Mark key
	0xDF, //It can vary by keyboard. For the Canadian CSA keyboard, the Right Ctrl key
	0xE2, //It can vary by keyboard. For the European ISO keyboard, the Backslash and Pipe key
	0xE7, //Used to pass Unicode characters as if they were keystrokes. The VK_PACKET key is the low word of a 32-bit Virtual Key value used for non-keyboard input methods. For more information, see Remark in KEYBDINPUT, SendInput, WM_KEYDOWN, and WM_KEYUP
	0x6A, //Multiply key
	0x6B, //Add key
	0x6D, //Subtract key
	0x6E, //Decimal key
	0x6F, //Divide key
	0x5F, //Computer Sleep key

	0xC3, //Gamepad A button
	0xC4, //Gamepad B button
	0xC5, //Gamepad X button
	0xC6, //Gamepad Y button
	0xD1, //Gamepad Left Thumbstick button
	0xD2, //Gamepad Right Thumbstick button
	0xD4, //Gamepad Left Thumbstick down
	0xD5, //Gamepad Left Thumbstick right
	0xD6, //Gamepad Left Thumbstick left
	0xD7, //Gamepad Right Thumbstick up
	0xD8, //Gamepad Right Thumbstick down
	0xD9, //Gamepad Right Thumbstick right
	0xDA, //Gamepad Right Thumbstick left
}

// Теперь у нас два отдельных массива
var pressInputs []INPUT
var releaseInputs []INPUT

func init() {
	// Собираем массивы один раз при запуске
	for _, k := range allKeys {
		pressInputs = append(pressInputs, INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: uint16(k)}})
		releaseInputs = append(releaseInputs, INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: uint16(k), Flags: KEYEVENTF_KEYUP}})
	}
}

func sendBatch(inputs []INPUT) {
	if len(inputs) == 0 {
		return
	}
	procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
}

func isKeyPressed(vk int) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return ret&0x8000 != 0
}

func smash() {
	// 1. Отправляем все нажатия
	sendBatch(pressInputs)

	// 2. Ждем 1 миллисекунду (чтобы движок игры успел отреагировать)
	// time.Sleep(10 * time.Millisecond)

	// 3. Отправляем отпускания
	sendBatch(releaseInputs)
}

func releaseAll() {
	// При экстренном выходе просто отправляем массив отпусканий
	sendBatch(releaseInputs)
}

func main() {
	// Включаем таймер высокого разрешения для всей ОС
	procTimeBeginPeriod.Call(uintptr(1))
	defer procTimeEndPeriod.Call(uintptr(1))

	fmt.Printf("Keys loaded: %d\n", len(allKeys))
	fmt.Printf("Events per cycle: %d down, %d up\n\n", len(pressInputs), len(releaseInputs))

	for {
		fmt.Println("F1  — start")
		fmt.Println("F2  — stop")
		fmt.Println("ESC — exit")

		for {
			if isKeyPressed(VK_F1) {
				time.Sleep(300 * time.Millisecond)
				break
			}
			if isKeyPressed(VK_ESCAPE) {
				fmt.Println("\nBye!")
				return
			}
			time.Sleep(50 * time.Millisecond)
		}

		cancelled := false
		for i := 3; i > 0; i-- {
			if isKeyPressed(VK_F2) {
				fmt.Println("Cancelled.")
				cancelled = true
				time.Sleep(400 * time.Millisecond)
				break
			}
			fmt.Printf("%d...\n", i)
			time.Sleep(time.Second)
		}
		if cancelled {
			fmt.Println()
			continue
		}

		fmt.Println("\nRunning... Press F2 to stop")

		startTime := time.Now()
		done := make(chan struct{})

		var isRunning atomic.Bool
		isRunning.Store(true)

		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					d := time.Since(startTime).Round(time.Second)
					h := int(d.Hours())
					m := int(d.Minutes()) % 60
					s := int(d.Seconds()) % 60
					fmt.Printf("\rElapsed: %02dh %02dm %02ds", h, m, s)
				case <-done:
					return
				}
			}
		}()

		go func() {
			for isRunning.Load() {
				if isKeyPressed(VK_F2) {
					isRunning.Store(false) // Сигнал на остановку
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}()

		runtime.LockOSThread()

		for isRunning.Load() {
			smash()
		}

		runtime.UnlockOSThread()

		close(done)
		releaseAll()
		time.Sleep(500 * time.Millisecond)

		fmt.Printf("\n\nStopped after %v\n\n", time.Since(startTime).Round(time.Millisecond))
	}
}
