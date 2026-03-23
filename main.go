package main

import (
	"fmt"
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
	"unicode"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	procSendInput       = user32.NewProc("SendInput")
	winmm               = syscall.NewLazyDLL("winmm.dll")
	procTimeBeginPeriod = winmm.NewProc("timeBeginPeriod")
	procTimeEndPeriod   = winmm.NewProc("timeEndPeriod")
)

const (
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
		Keys:    rangeKeys(0x70, 0x8F),
		Default: true,
	},
	{
		Name:    "OEM",
		Keys:    append([]byte{0xE1, 0xE3, 0xE4, 0xE6}, append(rangeKeys(0xE9, 0xF5), rangeKeys(0x92, 0x96)...)...), //0x92 work as two keys
		Default: true,
	},
	{
		Name:    "Modifier",
		Keys:    []byte{0x14, 0x90, 0xA0, 0xA1},
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
	// if len(inputs) == 0 {
	// 	return
	// }
	procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
}

func makeDelayEntry(defaultVal string) *widget.Entry {
	e := widget.NewEntry()
	e.SetText(defaultVal)
	e.OnChanged = func(s string) {
		filtered := ""
		for _, c := range s {
			if unicode.IsDigit(c) {
				filtered += string(c)
			}
		}
		if len(filtered) > 3 {
			filtered = filtered[:3]
		}
		if filtered != s {
			e.SetText(filtered)
		}
	}
	return e
}

func parseDelay(e *widget.Entry, fallback int) int {
	v, err := strconv.Atoi(e.Text)
	if err != nil || v < 0 {
		return fallback
	}
	if v > 999 {
		return 999
	}
	return v
}

func main() {
	procTimeBeginPeriod.Call(uintptr(1))
	defer procTimeEndPeriod.Call(uintptr(1))

	a := app.New()
	w := a.NewWindow("Keyboard Smasher")
	w.SetFixedSize(true)

	// --- Key group checkboxes ---
	selected := make([]bool, len(keyGroups))
	for i, g := range keyGroups {
		selected[i] = g.Default
	}

	totalLabel := widget.NewLabel("")
	updateTotal := func() {
		n := 0
		for i, g := range keyGroups {
			if selected[i] {
				n += len(g.Keys)
			}
		}
		totalLabel.SetText(fmt.Sprintf("Selected: %d keys", n))
	}

	checks := make([]fyne.CanvasObject, len(keyGroups))
	for i, g := range keyGroups {
		idx := i
		label := g.Name
		check := widget.NewCheck(label, func(v bool) {
			selected[idx] = v
			updateTotal()
		})
		check.Checked = selected[i]
		checks[i] = check
	}
	updateTotal()

	keyGroupsCard := widget.NewCard("Key Groups", "",
		container.NewVBox(
			container.NewGridWithColumns(2, checks...),
			totalLabel,
		),
	)

	// --- Delay inputs ---
	pressDelayEntry := makeDelayEntry("500")
	releaseDelayEntry := makeDelayEntry("500")

	noTiming := false
	noTimingCheck := widget.NewCheck("No delay", func(v bool) {
		noTiming = v
		if v {
			pressDelayEntry.Disable()
			releaseDelayEntry.Disable()
		} else {
			pressDelayEntry.Enable()
			releaseDelayEntry.Enable()
		}
	})

	timingCard := widget.NewCard("Timing", "", container.NewVBox(
		noTimingCheck,
		container.NewHBox(
			widget.NewLabel("Press:"),
			container.NewGridWrap(fyne.NewSize(45, 36), pressDelayEntry),
			widget.NewLabel("ms"),
			widget.NewLabel("   "),
			widget.NewLabel("Release:"),
			container.NewGridWrap(fyne.NewSize(45, 36), releaseDelayEntry),
			widget.NewLabel("ms"),
		),
	))

	dotRich := widget.NewRichText(&widget.TextSegment{
		Style: widget.RichTextStyle{ColorName: theme.ColorNameDisabled},
		Text:  "●",
	})
	textRich := widget.NewRichText(&widget.TextSegment{
		Style: widget.RichTextStyle{ColorName: theme.ColorNameForeground},
		Text:  "Ready",
	})
	setStatus := func(col fyne.ThemeColorName, text string) {
		dotRich.Segments = []widget.RichTextSegment{
			&widget.TextSegment{Style: widget.RichTextStyle{ColorName: col}, Text: "●"},
		}
		dotRich.Refresh()
		textRich.Segments = []widget.RichTextSegment{
			&widget.TextSegment{Style: widget.RichTextStyle{ColorName: theme.ColorNameForeground}, Text: text},
		}
		textRich.Refresh()
	}
	statusBar := container.NewPadded(container.NewHBox(dotRich, textRich))

	// --- Button ---
	var isRunning atomic.Bool
	var stopCh chan struct{}

	btn := widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), nil)
	btn.Importance = widget.SuccessImportance

	btn.OnTapped = func() {
		if isRunning.Load() {
			// Stop
			isRunning.Store(false)
			return
		}

		// Start
		if pressDelayEntry.Text == "" {
			pressDelayEntry.SetText("0")
		}
		if releaseDelayEntry.Text == "" {
			releaseDelayEntry.SetText("0")
		}
		pressMs := parseDelay(pressDelayEntry, 0)
		releaseMs := parseDelay(releaseDelayEntry, 0)

		press, release := buildInputs(selected)
		if len(press) == 0 {
			setStatus(theme.ColorNameError, "No keys selected!")
			return
		}
		combined := append(press, release...)
		useSingleBatch := noTiming

		btn.SetIcon(theme.MediaStopIcon())
		btn.SetText("Stop")
		btn.Importance = widget.DangerImportance
		btn.Refresh()
		isRunning.Store(true)
		stopCh = make(chan struct{})

		// Smash goroutine (с обратным отсчётом)
		go func() {
			for i := 3; i >= 1; i-- {
				if !isRunning.Load() {
					close(stopCh)
					fyne.Do(func() {
						setStatus(theme.ColorNameDisabled, "Cancelled")
						btn.SetIcon(theme.MediaPlayIcon())
						btn.SetText("Start")
						btn.Importance = widget.SuccessImportance
						btn.Refresh()
					})
					return
				}
				n := i
				fyne.Do(func() { setStatus(theme.ColorNameWarning, fmt.Sprintf("Starting in %d...", n)) })
				time.Sleep(1000 * time.Millisecond)
			}

			startTime := time.Now()

			// Timer goroutine стартует только после отсчёта
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
						txt := fmt.Sprintf("Running  %02dh %02dm %02ds  |  %d keys", h, m, s, len(press))
						fyne.Do(func() { setStatus(theme.ColorNameSuccess, txt) })
					case <-stopCh:
						elapsed := time.Since(startTime).Round(time.Second)
						fyne.Do(func() {
							setStatus(theme.ColorNameWarning, fmt.Sprintf("Stopped after %v", elapsed))
							btn.SetIcon(theme.MediaPlayIcon())
							btn.SetText("Start")
							btn.Importance = widget.SuccessImportance
							btn.Refresh()
						})
						return
					}
				}
			}()

			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			if useSingleBatch {
				for isRunning.Load() {
					sendBatch(combined)
					runtime.Gosched()
				}
			} else {
				for isRunning.Load() {
					sendBatch(press)
					time.Sleep(time.Duration(pressMs) * time.Millisecond)
					sendBatch(release)
					time.Sleep(time.Duration(releaseMs) * time.Millisecond)
				}
			}
			sendBatch(release)
			close(stopCh)
		}()
	}

	bottom := container.NewVBox(
		container.New(layout.NewCustomPaddedLayout(10, 10, 0, 0), btn),
		widget.NewSeparator(),
		statusBar,
	)

	content := container.NewBorder(
		nil, bottom, nil, nil,
		container.NewVBox(keyGroupsCard, timingCard),
	)

	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(480, 560))
	// w.Canvas().SetOnTypedKey(func(*fyne.KeyEvent) {})
	w.ShowAndRun()
}
