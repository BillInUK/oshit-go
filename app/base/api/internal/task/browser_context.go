package task

import (
	"context"
	"github.com/chromedp/chromedp"
	"time"
)

type BrowserContext struct {
	AllocCtx    context.Context
	AllocCancel context.CancelFunc
	Ctx         context.Context
	CtxCancel   context.CancelFunc
}

// CreateBrowserContext 创建浏览器上下文资源
func CreateBrowserContext(headless bool) (*BrowserContext, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless), // 调试时可设为false
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, ctxCancel := chromedp.NewContext(allocCtx)

	// 设置全局超时
	ctx, ctxCancel = context.WithTimeout(ctx, 1*time.Minute)

	return &BrowserContext{
		AllocCtx:    allocCtx,
		AllocCancel: allocCancel,
		Ctx:         ctx,
		CtxCancel:   ctxCancel,
	}, nil
}

func (ctx BrowserContext) Close() {
	ctx.CtxCancel()
	ctx.AllocCancel()
}
