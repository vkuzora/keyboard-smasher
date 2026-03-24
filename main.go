package main

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed assets/logo.png
var logoBytes []byte

var logoResource = fyne.NewStaticResource("logo.png", logoBytes)

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

	os.Setenv("FYNE_SCALE", "0.8")
	a := app.New()
	a.SetIcon(logoResource)
	w := a.NewWindow("Keyboard Smasher")
	w.SetFixedSize(true)
	w.SetIcon(logoResource)

	// --- Key group checkboxes ---
	selected := make([]bool, len(keyGroups))
	for i, g := range keyGroups {
		selected[i] = g.Default
	}

	if cfg := loadConfig(); cfg != nil {
		if len(cfg.Selected) == len(keyGroups) {
			copy(selected, cfg.Selected)
		}
	}

	var save func()

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
		check := widget.NewCheck(g.Name, func(v bool) {
			selected[idx] = v
			updateTotal()
			save()
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

	save = func() {
		saveConfig(Config{
			Selected:  selected,
			PressMs:   pressDelayEntry.Text,
			ReleaseMs: releaseDelayEntry.Text,
			NoTiming:  noTiming,
		})
	}

	if cfg := loadConfig(); cfg != nil {
		if cfg.PressMs != "" {
			pressDelayEntry.SetText(cfg.PressMs)
		}
		if cfg.ReleaseMs != "" {
			releaseDelayEntry.SetText(cfg.ReleaseMs)
		}
		noTiming = cfg.NoTiming
	}

	pressDelayEntry.OnChanged = func(s string) { save() }
	releaseDelayEntry.OnChanged = func(s string) { save() }

	noTimingCheck := widget.NewCheck("No delay", func(v bool) {
		noTiming = v
		if v {
			pressDelayEntry.Disable()
			releaseDelayEntry.Disable()
		} else {
			pressDelayEntry.Enable()
			releaseDelayEntry.Enable()
		}
		save()
	})
	noTimingCheck.Checked = noTiming
	if noTiming {
		pressDelayEntry.Disable()
		releaseDelayEntry.Disable()
	}

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

	// --- Status ---
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

	btn := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), nil)
	btn.Importance = widget.SuccessImportance

	btn.OnTapped = func() {
		if isRunning.Load() {
			isRunning.Store(false)
			return
		}

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
		btn.Importance = widget.DangerImportance
		btn.Refresh()
		isRunning.Store(true)
		stopCh = make(chan struct{})

		// F1 watcher
		go func() {
			for isRunning.Load() {
				if isKeyPressed(VK_F1) {
					isRunning.Store(false)
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}()

		// Focus watcher
		go func() {
			time.Sleep(500 * time.Millisecond)
			for isRunning.Load() {
				if !isForeground() {
					isRunning.Store(false)
					break
				}
				time.Sleep(50 * time.Millisecond)
			}
		}()

		// Smash goroutine
		go func() {
			for i := 3; i >= 1; i-- {
				if !isRunning.Load() {
					close(stopCh)
					fyne.Do(func() {
						setStatus(theme.ColorNameDisabled, "Cancelled")
						btn.SetIcon(theme.MediaPlayIcon())
						btn.Importance = widget.SuccessImportance
						btn.Refresh()
					})
					return
				}
				n := i
				fyne.Do(func() { setStatus(theme.ColorNameWarning, fmt.Sprintf("Starting in %d...", n)) })
				time.Sleep(time.Second)
			}

			startTime := time.Now()

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

	// --- Help ---
	var helpWin fyne.Window
	helpBtn := widget.NewButton("?", func() {
		if helpWin != nil {
			helpWin.RequestFocus()
			return
		}

		hw := a.NewWindow("Help")
		hw.SetFixedSize(true)
		hw.Resize(fyne.NewSize(450, 500))
		hw.SetOnClosed(func() { helpWin = nil })

		richText := widget.NewRichTextFromMarkdown(helpTexts["RU"])
		richText.Wrapping = fyne.TextWrapWord

		langGroup := widget.NewRadioGroup([]string{"RU", "EN"}, func(lang string) {
			richText.ParseMarkdown(helpTexts[lang])
		})
		langGroup.SetSelected("EN")
		langGroup.Horizontal = true

		hw.SetContent(container.NewPadded(container.NewBorder(
			container.NewHBox(layout.NewSpacer(), langGroup),
			nil, nil, nil,
			container.NewVScroll(richText),
		)))
		hw.SetIcon(theme.QuestionIcon())
		hw.CenterOnScreen()
		hw.Show()
		helpWin = hw
	})
	helpBtn.Importance = widget.LowImportance

	bottom := container.NewVBox(
		container.New(layout.NewCustomPaddedLayout(10, 10, 0, 0), btn),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, nil, helpBtn, statusBar),
	)

	content := container.NewBorder(
		nil, bottom, nil, nil,
		container.NewVBox(keyGroupsCard, timingCard),
	)

	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(480, 560))
	w.CenterOnScreen()
	w.ShowAndRun()
}
