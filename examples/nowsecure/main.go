package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/jarylc/go-chromedpproxy"
	"log"
)

func main() {
	// prepare proxy using port 9222 as Chrome debugging port and 9221 as frontend port
	// also send in `chromedp.DisableGPU` as extra ExecAllocatorOption
	chromedpproxy.PrepareProxy(":9222", ":9221", chromedp.DisableGPU)
	targetID, err := chromedpproxy.NewTab("https://nowsecure.nl", chromedp.WithLogf(log.Printf))
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Panic(err)
	}
	ctx := chromedpproxy.GetTarget(targetID)
	if err := chromedp.Run(ctx,
		// Check if we pass anti-bot measures.
		chromedp.Navigate("https://nowsecure.nl"),
		chromedp.WaitVisible(`//div[@class="hystericalbg"]`),
	); err != nil {
		log.Panic(err)
	}
	fmt.Println("Undetected!")
}
