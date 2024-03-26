package progressbar

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"strings"
	"sync"

	"github.com/fatih/color"
	mpb "github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"github.com/vend/govend/vend"
	"golang.org/x/crypto/ssh/terminal"
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
	DEFAULT_BAR_WIDTH   = 80
	DEFAULT_NAME_LENGTH = MEDIUM_NAME
	BUFFER              = 20 // arbitrary buffer to add to the bar width
)

type ProgressBar struct {
	Progress     *mpb.Progress
	NameLength   int
	BarWidth     int
	Defaults     *Defaults
	VendClient   *vend.Client
	WaitGroup    *sync.WaitGroup
	ErrorChannel chan error
	DataChannel  chan interface{}
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
	return fmt.Sprintf(" %s | %sCOMPLETE %s", name, GREEN, RESET) // green
}

func (d *Defaults) ErrorName(name string) string {
	return fmt.Sprintf(" %s | %s  ERROR  %s", name, RED, RESET) // red
}

func setBarWidth() (int, error) {
	terminalWidth, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0, errors.New("error getting terminal size")
	}

	default_total_width := DEFAULT_BAR_WIDTH + DEFAULT_NAME_LENGTH + BUFFER

	switch {
	case terminalWidth > default_total_width:
		return DEFAULT_BAR_WIDTH, nil
	case terminalWidth < default_total_width:
		adjustedWidth := terminalWidth - DEFAULT_NAME_LENGTH - BUFFER
		if adjustedWidth > 0 {
			return adjustedWidth, nil
		} else {
			return 0, errors.New("terminal too small for progress bar")
		}
	default:
		return DEFAULT_BAR_WIDTH, nil
	}
}

func (p *ProgressBar) setCompleteBarWidth() string {
	var completeBar = "["
	for i := 0; i < p.BarWidth-2; i++ { // -2 to account for the brackets
		completeBar += "🁢"
	}
	completeBar += "]"
	return completeBar
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

// creates a wait group and progress bar group
func CreateMultiBarGroup(numBars int, token string, domain string) (*ProgressBar, error) {
	width, err := setBarWidth()
	if err != nil {
		return &ProgressBar{}, err
	}
	var vc vend.Client

	var wg sync.WaitGroup
	if token != "" {
		vc = vend.NewClient(token, domain, "")
	}

	// each fetchData has a data1 and data2, so we need a buffer that is 2x the num of routines
	dataChannel := make(chan interface{}, numBars*2)
	errChannel := make(chan error, numBars)

	group := ProgressBar{
		Progress:     mpb.New(mpb.WithWaitGroup(&wg), mpb.WithWidth(width)),
		BarWidth:     width,
		NameLength:   DEFAULT_NAME_LENGTH,
		WaitGroup:    &wg,
		VendClient:   &vc,
		ErrorChannel: errChannel,
		DataChannel:  dataChannel,
	}
	return &group, nil
}

func CreateSingleBar() *ProgressBar {
	width, _ := setBarWidth()
	return &ProgressBar{Progress: mpb.New(mpb.WithWidth(width)), BarWidth: width}
}

func (p *ProgressBar) AddProgressBar(total int, name string) (*CustomBar, error) {
	style := p.Defaults.CreateBarStyle("[", "]", "🁢", "", "_", CYAN) // cyan
	return p.AddBarWithOptions(style, total, name, DEFAULT_NAME_LENGTH)
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
	name, err = setNameLength(name, DEFAULT_NAME_LENGTH)
	if err != nil {
		return &CustomBar{Bar: bar}, errors.New(fmt.Sprintf("set name length error: %s", err))
	}

	bar = p.Progress.New(int64(-1), style,
		mpb.BarFillerOnComplete(fmt.Sprintf("%s%s%s", CYAN, p.setCompleteBarWidth(), RESET)), // p.setCompleteBarWidth()
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

func (p *ProgressBar) fetchData(name string) {
	defer p.WaitGroup.Done()
	vc := *p.VendClient
	bar, err := p.AddIndeterminateProgressBar(name)
	if err != nil {
		p.ErrorChannel <- err
		return
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	var data interface{}
	var data2 interface{}

	switch name {
	case "products":
		data, data2, err = vc.Products()
	case "outlets":
		data, data2, err = vc.Outlets()
	case "outlet-taxes":
		data, err = vc.OutletTaxes()
	case "taxes":
		data, data2, err = vc.Taxes()
	case "inventory":
		data, err = vc.Inventory()
	case "product-tags":
		data, err = vc.Tags()
	case "registers":
		data, err = vc.Registers()
	case "users":
		data, err = vc.Users()
	case "user":
		data, err = vc.User()
	case "customers":
		data, err = vc.Customers()
	case "customer-groups":
		data, err = vc.CustomerGroups()
	case "store-credits":
		data, err = vc.StoreCredits()
	case "gift-cards":
		data, err = vc.GiftCards()
	}

	close(done)

	if err != nil {
		bar.AbortBar()
		p.ErrorChannel <- err
	} else {
		p.DataChannel <- data
		p.DataChannel <- data2
	}
	bar.SetIndeterminateBarComplete()
}

func (p *ProgressBar) FetchDataWithProgressBar(name string) {
	p.WaitGroup.Add(1)
	go p.fetchData(name)
}

func (p *ProgressBar) FetchSalesDataWithProgressBar(versionAfter int64) {
	p.WaitGroup.Add(1)
	go p.fetchSalesData(versionAfter)
}

func (p *ProgressBar) fetchSalesData(versionAfter int64) {
	defer p.WaitGroup.Done()
	vc := *p.VendClient
	bar, err := p.AddIndeterminateProgressBar("sales")
	if err != nil {
		p.ErrorChannel <- err
		return
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	data, err := vc.SalesAfter(versionAfter)

	close(done)

	if err != nil {
		bar.AbortBar()
		p.ErrorChannel <- err
	} else {
		p.DataChannel <- data
	}
	bar.SetIndeterminateBarComplete()
}

type Task func(args ...interface{}) interface{}

// perform a task with an indeterminate progress bar, "task" is a function with a single return
func (p *ProgressBar) PerformTaskWithProgressBar(barName string, task Task, args ...interface{}) {
	p.WaitGroup.Add(1)
	go p.performTask(barName, task, args...)
}

func (p *ProgressBar) performTask(barName string, task Task, args ...interface{}) {
	defer p.WaitGroup.Done()
	bar, err := p.AddIndeterminateProgressBar(barName)
	if err != nil {
		p.ErrorChannel <- err
		return
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	p.DataChannel <- task(args...)
	close(done)
	bar.SetIndeterminateBarComplete()
}

func (p *ProgressBar) MultiBarGroupWait() {
	p.WaitGroup.Wait()
	p.Progress.Wait()
	close(p.ErrorChannel)
	close(p.DataChannel)
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

func (bar *CustomBar) IncBy(amount int) {
	bar.Bar.IncrBy(amount)
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
