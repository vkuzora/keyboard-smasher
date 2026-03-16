package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procSendInput        = user32.NewProc("SendInput")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
)

const (
	INPUT_KEYBOARD  = 1
	KEYEVENTF_KEYUP = 0x0002
	VK_ESCAPE       = 0x1B
	VK_F8           = 0x77
	VK_F9           = 0x78
)

type KEYBDINPUT struct {
	Vk        uint16
	Scan      uint16
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

// INPUT mirrors the Win32 INPUT union (40 bytes) with explicit padding for 64-bit alignment.
type INPUT struct {
	Type uint32
	_    uint32
	Ki   KEYBDINPUT
	_    uint64
}

// 0xA0,
// 0xA1,
// 0xA2,
// 0xA3,
// 0xA4,
// 0xA5,
// 0x10,
// 0x11,
// 0x12,
// 0x14,
// 0xCB,
// 0xD3,

var allKeys = []byte{

	// 0x01,
	// 0x02,
	// 0x03,
	// 0x04,
	// 0x05,
	// 0x06,
	// 0x08,
	// 0x09,
	// 0x0C,
	// 0x0D,
	// 0x13,
	0x15,
	0x17,
	0x18,
	0x19,
	0x1C,
	0x1D,
	0x1E,
	0x1F,
	// 0x20,
	// 0x21,
	// 0x22,
	0x23,
	0x24,
	0x25,
	0x26,
	0x27,
	0x28,
	// 0x29,
	0x2B,
	0x2D,
	0x2E,
	0x2F,
	// 0x30,
	// 0x31,
	// 0x32,
	// 0x33,
	// 0x34,
	// 0x35,
	// 0x36,
	// 0x37,
	// 0x38,
	// 0x39,
	// 0x41,
	// 0x42,
	// 0x43,
	// 0x44,
	// 0x45,
	// 0x46,
	// 0x47,
	// 0x48,
	// 0x49,
	// 0x4A,
	// 0x4B,
	// 0x4C,
	// 0x4D,
	// 0x4E,
	// 0x4F,
	// 0x50,
	// 0x51,
	// 0x52,
	// 0x53,
	// 0x54,
	// 0x55,
	// 0x56,
	// 0x57,
	// 0x58,
	// 0x59,
	// 0x5A,
	0x5F,
	// 0x60,
	// 0x61,
	// 0x62,
	// 0x63,
	// 0x64,
	// 0x65,
	// 0x66,
	// 0x67,
	// 0x68,
	// 0x69,
	0x6A,
	0x6B,
	0x6D,
	0x6E,
	0x6F,
	// 0x70,
	// 0x71,
	// 0x72,
	// 0x73,
	// 0x74,
	// 0x75,
	// 0x76,
	// 0x77,
	// 0x78,
	// 0x79,
	// 0x7B,
	// 0x7C,
	// 0x7D,
	// 0x7E,
	// 0x7F,
	// 0x80,
	// 0x81,
	// 0x82,
	// 0x83,
	// 0x84,
	// 0x85,
	// 0x86,
	// 0x87,
	0x88,
	0x89,
	0x8A,
	0x8B,
	0x8C,
	0x8D,
	0x8E,
	0x8F,
	// 0x90,
	// 0x91,
	0x92,
	0x93,
	0x94,
	0x95,
	0x96,
	0xA6,
	0xA7,
	0xA8,
	0xA9,
	0xBA,
	0xBB,
	0xBC,
	0xBD,
	0xBE,
	0xBF,
	0xC0,
	0xC3,
	0xC4,
	0xC5,
	0xC6,
	0xC7,
	0xC8,
	0xC9,
	0xCA,
	0xCC,
	0xCD,
	0xCE,
	0xCF,
	0xD0,
	0xD1,
	0xD2,
	0xD4,
	0xD5,
	0xD6,
	0xD7,
	0xD8,
	0xD9,
	0xDA,
	0xDB,
	0xDC,
	0xDD,
	0xDE,
	0xDF,
	0xE1,
	0xE2,
	0xE3,
	0xE4,
	0xE5,
	0xE6,
	0xE7,
	0xE9,
	0xEA,
	0xEB,
	0xEC,
	0xED,
	0xEE,
	0xEF,
	0xF0,
	0xF1,
	0xF2,
	0xF3,
	0xF4,
	0xF5,
	0xF6,
	0xF7,
	0xF8,
	0xF9,
	0xFA,
	0xFB,
	0xFC,
	0xFD,
	0xFE,
}

// pressInputs and releaseInputs are pre-built once at startup to avoid
// per-iteration allocations during loop.
var pressInputs, releaseInputs []INPUT

// init builds the press and release INPUT arrays for every key in allKeys.
// Both arrays are populated in a single pass to keep cache behaviour symmetric.
func init() {
	for _, k := range allKeys {
		pressInputs = append(pressInputs, INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: uint16(k)}})
		releaseInputs = append(releaseInputs, INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: uint16(k), Flags: KEYEVENTF_KEYUP}})
	}
}

// sendBatch dispatches all inputs in a single SendInput syscall.
// Sending the whole slice at once is faster than one call per key and
// makes the key events appear simultaneous to the OS event queue.
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

// isKeyPressed returns true if the virtual key vk is currently held down.
// The high-order bit of GetAsyncKeyState's return value indicates key state.
func isKeyPressed(vk int) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return ret&0x8000 != 0
}

// presses and immediately releases all keys in a single cycle.
func smash() {
	sendBatch(pressInputs)
	// time.Sleep(10 * time.Millisecond)
	sendBatch(releaseInputs)
	// time.Sleep(10 * time.Millisecond)

}

// releaseAll sends a release event for every key to ensure no keys remain
// stuck in the pressed state after loop exits.
func releaseAll() { sendBatch(releaseInputs) }

// main is the entry point and control loop.
// Each iteration waits for F9 or ESC, runs a 5-second countdown,
// then smashes all keys until ESC is pressed.
func main() {
	fmt.Printf("Keys loaded: %d\n\n", len(allKeys))

	for {
		fmt.Println("F8       — start")
		fmt.Println("Hold F9  — stop")
		fmt.Println("ESC      — exit")

		// Poll until the user picks an action.
		for {
			if isKeyPressed(VK_F8) {
				// Brief sleep to avoid re-triggering on key hold.
				time.Sleep(300 * time.Millisecond)
				break
			}
			if isKeyPressed(VK_ESCAPE) {
				fmt.Println("\nBye!")
				return
			}
			time.Sleep(50 * time.Millisecond)
		}

		// Give the user time to switch focus to the target window.
		cancelled := false
		for i := 3; i > 0; i-- {
			if isKeyPressed(VK_F9) {
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

		fmt.Println("\nRunning — hold F9 to stop")

		startTime := time.Now()
		done := make(chan struct{})

		// Background goroutine prints a live elapsed timer without blocking loop.
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

		for !isKeyPressed(VK_F9) {
			smash()
		}

		// Signal the timer goroutine to stop, then release all keys before exiting.
		close(done)
		releaseAll()
		time.Sleep(500 * time.Millisecond)

		fmt.Printf("\n\nStopped after %v\n\n", time.Since(startTime).Round(time.Millisecond))
	}
}
