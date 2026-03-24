package main

var helpTexts = map[string]string{
	"RU": `## Key Groups
Список групп клавиш. Каждая группа содержит своё количество клавиш. Выбранные группы будут задействованы при запуске.

---

## Timing
**No delay** — все нажатия и отжатия отправляются одним пакетом без задержек. Максимальная скорость.

**Press / Release** — задержка в миллисекундах между нажатием и отжатием клавиш.

---

## Кнопка старт / стоп
Запускает и останавливает работу скрипта. В случае, если кнопка остановки не прожимается, можно нажать **F1** — это остановит скрипт независимо от состояния интерфейса.

---

## Статус и таймер
Нижняя строка показывает текущее состояние и время работы. При очень высокой скорости интерфейс может временно подвисать — это нормально, скрипт продолжает работать в фоне.`,

	"EN": `## Key Groups
A list of key groups. Each group contains a different number of keys. Selected groups will be used when the script runs.

---

## Timing
**No delay** — all key presses and releases are sent in a single batch with no delay. Maximum speed.

**Press / Release** — delay in milliseconds between pressing and releasing keys.

---

## Start / Stop button
Starts and stops the script. If the stop button is unresponsive, press **F1** — it will stop the script regardless of the UI state.

---

## Status and timer
The bottom bar shows the current state and elapsed time. At very high speeds the UI may temporarily freeze — this is normal, the script continues running in the background.`,
}
