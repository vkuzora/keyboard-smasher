package main

import "unsafe"

type KeyGroup struct {
	Name    string
	Keys    []byte
	Default bool
}

func rangeKeys(from, to byte) []byte {
	keys := make([]byte, 0, int(to-from)+1)
	for i := from; i <= to; i++ {
		keys = append(keys, i)
	}
	return keys
}

var keyGroups = []KeyGroup{
	{
		Name:    "Mouse",
		Keys:    rangeKeys(0x01, 0x06),
		Default: true,
	},
	{
		Name:    "Standard Control",
		Keys:    []byte{0x08, 0x0C, 0x0D, 0x13, 0x20}, //wo tab ui moment
		Default: true,
	},
	{
		Name:    "IME",
		Keys:    append([]byte{0x15, 0xE5}, append(rangeKeys(0x17, 0x19), rangeKeys(0x1C, 0x1F)...)...),
		Default: true,
	},
	{
		Name:    "Navigation",
		Keys:    rangeKeys(0x21, 0x24),
		Default: true,
	},
	{
		Name:    "Arrow",
		Keys:    rangeKeys(0x25, 0x28),
		Default: true,
	},

	{
		Name:    "Editing",
		Keys:    append([]byte{0x29, 0x2B}, rangeKeys(0x2D, 0x2E)...),
		Default: true,
	},
	{
		Name:    "Number",
		Keys:    rangeKeys(0x30, 0x39),
		Default: true,
	},
	{
		Name:    "Alphabet",
		Keys:    rangeKeys(0x41, 0x5A),
		Default: true,
	},

	{
		Name:    "Numpad",
		Keys:    rangeKeys(0x60, 0x6F),
		Default: true,
	},
	{
		Name:    "Function",
		Keys:    rangeKeys(0x71, 0x8F),
		Default: true,
	},
	{
		Name:    "OEM",
		Keys:    append([]byte{0xE1, 0xE3, 0xE4, 0xE6}, append(rangeKeys(0xE9, 0xF5), rangeKeys(0x92, 0x96)...)...), //0x92 work as two keys
		Default: true,
	},
	{
		Name:    "Browser",
		Keys:    rangeKeys(0xA6, 0xA9),
		Default: true,
	},
	{
		Name:    "Symbols",
		Keys:    append([]byte{0xE2}, append(rangeKeys(0xBA, 0xC0), rangeKeys(0xDB, 0xDF)...)...),
		Default: true,
	},
	{
		Name:    "Gamepad",
		Keys:    append(rangeKeys(0xD4, 0xDA), append(rangeKeys(0xC3, 0xCA), rangeKeys(0xCC, 0xD2)...)...),
		Default: false,
	},
	{
		Name:    "Other",
		Keys:    append([]byte{0xE7, 0x5F}, rangeKeys(0xF6, 0xFD)...),
		Default: true,
	},
}

func buildInputs(selected []bool) (press, release []INPUT) {
	for i, g := range keyGroups {
		if !selected[i] {
			continue
		}
		for _, k := range g.Keys {
			press = append(press, INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: uint16(k)}})
			release = append(release, INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: uint16(k), Flags: KEYEVENTF_KEYUP}})
		}
	}
	return
}

func sendBatch(inputs []INPUT) {
	procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
}
