package main

import (
	"context"
	"errors"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/jarylc/go-chromedpproxy"
	"log"
)

func main() {
	// create result channel to signify end of program
	resultChan := make(chan string)

	// prepare proxy using port 9222 as Chrome debugging port and 9221 as frontend port
	// also send in `chromedp.DisableGPU` as extra ExecAllocatorOption
	chromedpproxy.PrepareProxy(":9222", ":9221", chromedp.DisableGPU)
	targetID, err := chromedpproxy.NewTab("https://www.hcaptcha.com", chromedp.WithLogf(log.Printf))
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Panic(err)
	}
	defer chromedpproxy.CloseTarget(targetID)

	ctx := chromedpproxy.GetTarget(targetID)

	// grab hcaptcha iframe
	var iframes []*cdp.Node
	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Click(`#radio-5`, chromedp.NodeVisible),
		chromedp.Nodes(`.h-captcha > iframe`, &iframes),
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Panic(err)
	}

	// parallel steps to check if a captcha really exists before notifying
	go func() {
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Click(`#checkbox`, chromedp.NodeVisible, chromedp.FromNode(iframes[0])),
			chromedp.WaitVisible(`.challenge-container`, chromedp.FromNode(iframes[0])),
			chromedp.ActionFunc(func(ctx context.Context) error {
				log.Print("You would normally send this via any form of notifications to a user:")
				log.Printf("Recaptcha detected, please solve it here: http://127.0.0.1:9221/?id=%s", targetID)
				return nil
			}),
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Panic(err)
		}
	}()

	// parallel steps to find success condition
	go func() {
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.WaitVisible(`.check`, chromedp.FromNode(iframes[0])),
			chromedp.ActionFunc(func(ctx context.Context) error {
				log.Print("Captcha solved!")
				resultChan <- "OK"
				return nil
			}),
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Panic(err)
		}
		resultChan <- "Failed"
	}()

	// you can do other stuff here like grab cookie data, etc.

	// log and close
	log.Print("Result: ", <-resultChan)
}
