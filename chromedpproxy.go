package chromedpproxy

import (
	"context"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"strings"
	"sync"
)

var mutex = sync.RWMutex{}
var loaded = make(chan bool, 1)

var Context context.Context

// PrepareProxy abstracts chromedp.NewExecAllocator to the use case of this package
// it accepts listen addresses for both Chrome remote debugging and frontend as configuration
// it is also a variadic function that accepts extra chromedp.ExecAllocatorOption to be passed to the chromedp.NewExecAllocator
func PrepareProxy(chromeListenAddr string, frontendListenAddr string, customOpts ...chromedp.ExecAllocatorOption) {
	// ensure only exactly one context is prepared
	mutex.Lock()
	if Context != nil {
		mutex.Unlock()
		return
	}

	// split up chromeListenAddr, default host to 127.0.0.1 if not specified
	chromeListenAddrSplit := strings.Split(chromeListenAddr, ":")
	if chromeListenAddrSplit[0] == "" {
		chromeListenAddrSplit[0] = "127.0.0.1"
	}

	// insert remote-debugging flags and any additional options
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("remote-debugging-address", chromeListenAddrSplit[0]),
		chromedp.Flag("remote-debugging-port", chromeListenAddrSplit[1]),
	)
	if len(customOpts) > 0 {
		opts = append(opts, customOpts...)
	}

	// create context and keep alive
	go func() {
		ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancel()
		Context = ctx
		loaded <- true
		mutex.Unlock()
		startFrontEnd(frontendListenAddr, chromeListenAddrSplit[1])
	}()
}

// NewTab abstracts creating a new tab in the root context
// it returns a target ID or error
func NewTab(url string, customOpts ...chromedp.ContextOption) (target.ID, error) {
	// if context is not prepared, create with default values
	if Context == nil {
		PrepareProxy(":9222", ":9221")
		<-loaded
	}
	mutex.Lock()
	defer mutex.Unlock()

	// create new tab and navigate to URL
	Context, _ = chromedp.NewContext(Context, customOpts...)
	err := chromedp.Run(Context, chromedp.Tasks{
		chromedp.Navigate(url),
	})
	if err != nil {
		return "", err
	}

	// return target ID
	chromeContext := chromedp.FromContext(Context)
	return chromeContext.Target.TargetID, nil
}

// GetTarget returns a context from a target ID
func GetTarget(id target.ID) context.Context {
	mutex.RLock()
	defer mutex.RUnlock()

	// return context from target ID
	ctx, _ := chromedp.NewContext(Context, chromedp.WithTargetID(id))
	return ctx
}

// CloseTarget closes a target by closing the page
// it returns an error if any
func CloseTarget(id target.ID) error {
	mutex.Lock()
	defer mutex.Unlock()

	Context, _ = chromedp.NewContext(Context, chromedp.WithTargetID(id))
	if err := chromedp.Run(Context, page.Close()); err != nil {
		return err
	}
	return nil
}
