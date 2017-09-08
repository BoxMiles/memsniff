package presentation

import (
	"errors"
	"fmt"
	"github.com/box/memsniff/analysis"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"strconv"
	"time"
)

const (
	numColumns  = 12
	statusLines = 1
	logLines    = 4
)

var (
	errQuitRequested = errors.New("user requested to quit")
)

func (u *uiContext) runTermbox() error {
	err := termbox.Init()
	if err != nil {
		return err
	}
	defer func() {
		// ensure that the termboxEvents goroutine shuts down
		termbox.Interrupt()
		termbox.Close()
	}()

	return u.eventLoop()
}

func (u *uiContext) eventLoop() error {
	updateTick := time.NewTicker(u.interval)
	defer updateTick.Stop()
	events := termboxEvents()
	if err := u.update(); err != nil {
		return err
	}

	for {
		select {
		case <-updateTick.C:
			if err := u.update(); err != nil {
				return err
			}

		case msg := <-u.msgChan:
			u.handleNewMessage(msg)

		case ev := <-events:
			if err := u.handleEvent(ev); err != nil {
				if err == errQuitRequested {
					return nil
				}
				return err
			}
		}
	}
}

func termboxEvents() <-chan termbox.Event {
	ch := make(chan termbox.Event)
	go func() {
		for {
			ev := termbox.PollEvent()
			if ev.Type == termbox.EventInterrupt {
				break
			}
			ch <- ev
		}
	}()
	return ch
}

func (u *uiContext) handleEvent(ev termbox.Event) error {
	switch ev.Type {
	case termbox.EventKey:
		if ev.Ch == 'p' {
			u.handlePause()
		}
		if ev.Ch == 'q' || ev.Key == termbox.KeyCtrlC {
			return errQuitRequested
		}
		if ev.Key == termbox.KeyCtrlL {
			if err := u.update(); err != nil {
				return err
			}
			if err := termbox.Sync(); err != nil {
				return err
			}
		}

	case termbox.EventResize:
		if err := u.update(); err != nil {
			return err
		}
	}
	return nil
}

func (u *uiContext) handlePause() {
	u.paused = !u.paused
	if u.paused {
		u.Log("Updates paused")
	} else {
		u.Log("Updates unpaused")
	}
}

func (u *uiContext) handleNewMessage(msg string) {
	if len(u.messages) < logLines {
		u.messages = append(u.messages, msg)
	} else {
		u.messages = append(u.messages[1:], msg)
	}
}

func renderHeader() {
	renderText(0, 0, "Key")
	renderTextRight(8, 0, "Reads")
	renderTextRight(9, 0, "Size")
	renderTextRight(11, 0, "Bandwidth")
	renderLine(0, 12, 1, '-')
}

func renderReport(rep analysis.Report) {
	lastY := yFromBottom(statusLines + logLines)
	for i, kr := range rep.Keys {
		y := i + 2
		if y > lastY {
			break
		}
		leader := ' '
		if kr.VariableSize {
			leader = '~'
		}
		renderText(0, y, kr.Name)
		renderInt(8, y, kr.GetHits, ' ')
		renderInt(9, y, kr.Size, leader)
		renderInt(11, y, kr.TotalTraffic, ' ')
	}
}

func (u *uiContext) renderMessages() {
	for i, msg := range u.messages {
		renderText(0, yFromBottom(i+statusLines), msg)
	}
}

func (u *uiContext) renderFooter(rep analysis.Report) {
	y := yFromBottom(0)
	stats := u.statProvider()
	renderText(0, y, rep.Timestamp.Format("15:04:05.000"))

	renderText(1, y, dropLabel(stats))
	renderText(4, y, fmt.Sprintf("Packets: %10d", stats.PacketsPassedFilter))
	renderText(6, y, fmt.Sprintf("GET responses: %10d", stats.ResponsesParsed))
}

func dropLabel(s Stats) string {
	var dropRate float64
	if s.PacketsPassedFilter == 0 {
		dropRate = 0
	} else {
		dropRate = float64(s.PacketsDroppedTotal) / float64(s.PacketsPassedFilter)
	}

	return fmt.Sprintf("Dropped: %d+%d+%d=%d (%5.2f%%)",
		s.PacketsDroppedKernel, s.PacketsDroppedParser,
		s.PacketsDroppedAnalysis, s.PacketsDroppedTotal, dropRate*100)
}

func renderText(column int, y int, txt string) {
	renderTextX(columnX(column), y, txt)
}

func renderTextRight(column int, y int, txt string) {
	lastX := columnX(column + 1)
	renderTextX(lastX - len(txt), y, txt)
}

func renderTextX(x, y int, txt string) {
	runes := []rune(txt)

	for _, r := range runes {
		termbox.SetCell(x, y, r, termbox.ColorDefault, termbox.ColorDefault)
		x += runewidth.RuneWidth(r)
	}
}

func renderInt(column int, y int, n int, leader rune) {
	str := strconv.Itoa(n)
	lastX := columnX(column + 1)
	termbox.SetCell(lastX - len(str) - 1, y, leader, termbox.ColorDefault, termbox.ColorDefault)
	renderTextX(lastX - len(str), y, str)
}

func renderLine(column int, span int, y int, ch rune) {
	w := runewidth.RuneWidth(ch)
	for x := columnX(column); x < columnX(column+span); x += w {
		termbox.SetCell(x, y, ch, termbox.ColorDefault, termbox.ColorDefault)
	}
}

func columnX(col int) int {
	w, _ := termbox.Size()
	if col >= numColumns {
		return w
	}
	return w * col / numColumns
}

func yFromBottom(n int) int {
	_, h := termbox.Size()
	return h - 1 - n
}

func (u *uiContext) update() error {
	err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if err != nil {
		return err
	}

	// Continue to clear the accumulated data every interval even when paused
	// so we don't get a big burst of data on unpause.
	rep := u.analysis.Report(!u.cumulative)
	if !u.paused {
		u.prevReport = rep
	}
	renderHeader()
	renderReport(u.prevReport)
	u.renderFooter(u.prevReport)
	u.renderMessages()

	return termbox.Flush()
}
