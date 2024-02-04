package progressbar

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"strings"
	"sync"

	"github.com/fatih/color"
	mpb "github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// Name Lengths
const (
	SMALL_NAME  = 12
	MEDIUM_NAME = 24
	LARGE_NAME  = 36
)

// Colors
// NOTE: using terminal escape codes for colors. There is a bug in the mpb library
// with mpd.Width that prevents the use of color.Color
const (
	RED    = "\033[01;31m"
	GREEN  = "\033[01;32m"
	YELLOW = "\033[01;33m"
	BLUE   = "\033[01;34m"
	CYAN   = "\033[01;36m"
	RESET  = "\033[0m"
)

const (
	BAR_WIDTH = 80
)

const (
	// not sure how the relationship with this length is with the length of mpb.BarWidth is calculated
	// this matches a bar width of 80
	// TODO: figure out how to calculate the length of the bar based on set mp.BarWidth
	completeBar = "[🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢🁢]"
)

type ProgressBar struct {
	Progress   *mpb.Progress
	NameLength int
	Defaults   *Defaults
}

type CustomBar struct {
	Bar *mpb.Bar
}

type Defaults struct {
	Name string
}

func (d *Defaults) LoadingName(name string) string {
	return fmt.Sprintf(" %s | loading..", name)
}

func (d *Defaults) CompletedName(name string) string {
	return fmt.Sprintf(" %s | %scomplete %s", name, GREEN, RESET) // green
}

func (d *Defaults) ErrorName(name string) string {
	return fmt.Sprintf(" %s | %s  ERROR  %s", name, RED, RESET) // red
}

func (d *Defaults) CreateBarStyle(leftBound, rightBound, filler, tip, padding, color string) mpb.BarStyleComposer {
	return mpb.BarStyle().
		Lbound(leftBound).LboundMeta(func(s string) string {
		return color + s + RESET
	}).
		Rbound(rightBound).RboundMeta(func(s string) string {
		return color + s + RESET
	}).
		Filler(filler).FillerMeta(func(s string) string {
		return color + s + RESET
	}).
		Tip(tip).TipMeta(func(s string) string {
		return color + s + RESET
	}).
		TipOnComplete(). // leave tip on complete
		Padding(padding).PaddingMeta(func(s string) string {
		return color + s + RESET
	})
}

func CreateMultiBarGroup(wg *sync.WaitGroup) *ProgressBar {
	return &ProgressBar{Progress: mpb.New(mpb.WithWaitGroup(wg), mpb.WithWidth(BAR_WIDTH))}
}

func CreateSingleBar() *ProgressBar {
	return &ProgressBar{Progress: mpb.New(mpb.WithWidth(BAR_WIDTH))}
}

func (p *ProgressBar) AddProgressBar(total int, name string) (*CustomBar, error) {
	style := p.Defaults.CreateBarStyle("[", "]", "🁢", "", "_", CYAN) // cyan
	return p.AddBarWithOptions(style, total, name, MEDIUM_NAME)
}

func (p *ProgressBar) AddBarWithOptions(barStyle mpb.BarStyleComposer, total int, name string, nameLength int) (*CustomBar, error) {

	var bar *mpb.Bar
	var err error
	name, err = setNameLength(name, nameLength)

	if err != nil {
		return &CustomBar{Bar: bar}, errors.New(fmt.Sprintf("set name length error: %s", err))
	}

	bar = p.Progress.New(int64(total), barStyle,
		mpb.PrependDecorators(
			decor.OnAbort(decor.OnComplete(
				decor.Meta(decor.Spinner(nil, decor.WCSyncSpace), toMetaFunc(color.New(color.FgRed))), "✔"), "✘"),
			decor.OnAbort(decor.OnComplete(
				decor.Name(p.Defaults.LoadingName(name)), p.Defaults.CompletedName(name)), p.Defaults.ErrorName(name)),
		),
		mpb.AppendDecorators(
			decor.NewPercentage("%d", decor.WCSyncSpace),
		),
	)

	return &CustomBar{Bar: bar}, nil
}

func (p *ProgressBar) AddIndeterminateProgressBar(name string) (*CustomBar, error) {
	style := p.Defaults.CreateBarStyle("[", "]", "_", "🁢🁢🁢🁢🁢🁢🁢🁢", "_", CYAN) // cyan
	var bar *mpb.Bar
	var err error
	name, err = setNameLength(name, MEDIUM_NAME)

	if err != nil {
		return &CustomBar{Bar: bar}, errors.New(fmt.Sprintf("set name length error: %s", err))
	}

	bar = p.Progress.New(int64(-1), style,
		mpb.BarFillerOnComplete(fmt.Sprintf("%s%s%s", CYAN, completeBar, RESET)),
		mpb.PrependDecorators(
			decor.OnAbort(decor.OnComplete(
				decor.Meta(decor.Spinner(nil, decor.WCSyncSpace), toMetaFunc(color.New(color.FgRed))), "✔"), "✘"),
			decor.OnAbort(decor.OnComplete(
				decor.Name(p.Defaults.LoadingName(name)), p.Defaults.CompletedName(name)), p.Defaults.ErrorName(name)),
		),
		mpb.AppendDecorators(
			decor.Elapsed(decor.ET_STYLE_GO, decor.WCSyncWidth),
		),
	)

	bar.SetTotal(100, false)

	return &CustomBar{Bar: bar}, nil

}

func (p *ProgressBar) Wait() {
	p.Progress.Wait()
}

func (bar *CustomBar) iterateIndeterminateBar() {
	current := bar.Bar.Current()
	switch {
	case current == 0:
		bar.home()
		// 100 is arbitary and only matters compared to the iteration amount. 100:5 = 20 iterations for a full bar
	case current > 100:
		bar.home()
	default:
		bar.Bar.IncrBy(5)
	}
}

func (bar *CustomBar) home() {
	randomNumber := int64(rand.Intn(11)) // between 0 and 10
	bar.Bar.SetCurrent(randomNumber)
}

func (bar *CustomBar) Increment() {
	bar.Bar.Increment()
}

func (bar *CustomBar) AnimateIndeterminateBar(done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			bar.iterateIndeterminateBar()
			time.Sleep(100 * time.Millisecond) // .1 sec
		}
	}

}

func (bar *CustomBar) SetIndeterminateBarComplete() {
	if !bar.Bar.Aborted() {
		bar.Bar.SetTotal(-1, true)
	}
}

func (bar *CustomBar) AbortBar() {
	bar.Bar.Abort(false)
}

func setNameLength(input string, length int) (string, error) {
	input = " " + input // prepend a space

	if len(input) == length {
		// String is already the desired length, no action needed
		return input, nil
	} else if len(input) < length {
		// Add whitespace until the string is exactly the desired length
		input = input + strings.Repeat(" ", length-len(input))
		return input, nil
	} else if len(input) > length {
		// Trim the string if it is too long
		input = input[:length-3] + "..."
		return input, nil
	} else {
		// Return an error
		return "", errors.New("string length error")
	}
}

func toMetaFunc(c *color.Color) func(string) string {
	return func(s string) string {
		return c.Sprint(s)
	}
}
