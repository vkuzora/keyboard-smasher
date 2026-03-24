# Keyboard Smasher

🇷🇺 [Русский](README.ru.md)

## Quick Start

1. Download `Keyboard Smasher.exe` from the [Releases](../../releases/latest) page
2. Run the file
3. Select the key groups you want to use
4. Set press/release delays or enable **No delay** for maximum speed
5. Click **Start**

To stop — click the **Stop** button or press **F1**.

---

## Interface

### Key Groups

A list of key groups. Check the ones you want included. The **Selected** counter shows the total number of active keys.

| Group | Keys |
|---|---|
| Mouse | Mouse buttons |
| Standard Control | Backspace, Enter, Space, etc. |
| Navigation | Page Up/Down, Home, End |
| Arrow | Arrow keys |
| Editing | Select All, Undo, Insert, Delete |
| Number | Digits 0–9 |
| Alphabet | Letters A–Z |
| Numpad | Numeric keypad and operators |
| Function | F2–F24 (F1 is reserved for emergency stop) |
| Browser | Browser media/navigation keys |
| Symbols | `;  =  ,  -  .  /  ~` and brackets |
| IME | Japanese/Chinese input keys |
| OEM | Manufacturer-specific keys |
| Gamepad | Gamepad buttons |
| Other | Miscellaneous system keys |

### Timing

**No delay** — presses and releases are sent in a single batch with no delay. Maximum speed; delay fields are disabled.

**Press / Release** — delay in milliseconds between pressing and releasing keys. Range: 0–999 ms. Empty field defaults to 0.

### Start / Stop button

Click to start. A countdown runs before input begins — use this time to switch to the target window.

The button turns into a red **Stop** — click it to stop.

> **Emergency stop:** press **F1** at any time — the script will stop even if the UI is unresponsive.

### Status bar

The bottom bar shows the current state and a running timer:

- `Ready` — idle
- `Starting in 3...` — countdown
- `Running 00h 00m 12s | 128 keys` — active
- `Stopped after 12s` — finished

> In **No delay** mode or with very short intervals the status bar may temporarily freeze.

---

## Behaviour

- The script **stops automatically** if the application window loses focus.
- Settings (selected groups and timings) are **saved automatically** and restored on next launch.
- Config is stored at `%APPDATA%\keyboard-smasher\config.json`.

---

## Requirements

- Windows 10 / 11 (x64)
- No administrator rights required

---

## Build from source

```bash
go build -ldflags "-s -w -H windowsgui"
```
