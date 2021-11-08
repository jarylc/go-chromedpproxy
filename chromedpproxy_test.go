package chromedpproxy

import (
	"context"
	"errors"
	"github.com/chromedp/chromedp"
	"testing"
)

func TestChromeDPProxy(t *testing.T) {
	// prepare proxy
	PrepareProxy(":9222", ":9221")

	// test creating proxy, then deleting, then creating again
	targetID, err := NewTab("about:blank")
	if err != nil {
		t.Error(err)
	}
	err = CloseTarget(targetID)
	if err != nil {
		t.Error(err)
	}

	err = CloseTarget(targetID)
	t.Logf("expected error: %s", err) // should error if attempting to close again
	targetID, err = NewTab("https://github.com/jarylc/go-chromedpproxy")
	if err != nil {
		t.Error(err)
	}

	// test regular chromedp usage
	ctx := GetTarget(targetID)
	result := ""
	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.InnerHTML(`strong[itemprop="name"] > a`, &result, chromedp.NodeVisible),
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Error(err)
	}
	if result != "go-chromedpproxy" {
		t.Errorf("expected result to be 'go-chromedpproxy', got %s", result)
	}

	// cleanup
	err = CloseTarget(targetID)
	if err != nil {
		t.Error(err)
	}
}
