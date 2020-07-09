package server

import (
	"context"
	"github.com/chromedp/cdproto/dom"
	ch "github.com/cjp2600/assr/cache"
	"github.com/cjp2600/assr/config"
	"github.com/fasthttp/router"
	"github.com/patrickmn/go-cache"
	"github.com/valyala/fasthttp"
	"log"
	"sync"
	"time"

	//"time"

	"github.com/chromedp/chromedp"
	"strings"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (s *Parser) Run(dir string) error {
	r := router.New()
	r.NotFound = DefaultProxyHandler

	for _, route := range config.GetRoutes() {
		r.GET(route.Path, SsrHandler)
	}

	serv := fasthttp.Server{
		Handler:        r.Handler,
		ReadBufferSize: 100 * 1024,
	}

	return serv.ListenAndServe(":" + config.GetAppPort())
}

func PreloadOnStart() {
	var wg sync.WaitGroup
	wg.Add(len(config.GetRoutes()))
	for _, rt := range config.GetRoutes() {
		go func(wg *sync.WaitGroup, rt config.Routes) {
			defer wg.Done()
			_ = chromeReloadContent(rt.Path, rt.Timeout, "cache"+rt.Path)
		}(&wg, rt)
	}
	wg.Wait()
}

func SsrHandler(ctx *fasthttp.RequestCtx) {
	var (
		path    = string(ctx.Path())
		cacheId = "cache" + path
		res     string
		timeout = config.GetRouteTimeout(ctx.Path())
	)

	html, found := ch.GetClient().Get(cacheId)
	if found {
		res = html.(string)
		go func() {
			res = chromeReloadContent(path, timeout, cacheId)
		}()
	} else {
		res = chromeReloadContent(path, timeout, cacheId)
	}
	ctx.SetContentType("text/html; charset=utf-8")
	ctx.Response.SetBody([]byte(res))
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func chromeReloadContent(path string, timeout time.Duration, cacheId string) string {
	var res string
	var url = config.GetStaticDomain(path)
	var ops []chromedp.ContextOption
	if config.IsDebug() {
		ops = append(ops, chromedp.WithDebugf(log.Printf))
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true),
		chromedp.Flag("ignore-certificate-errors", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctxw, cancel := chromedp.NewContext(allocCtx, ops...)
	defer cancel()

	err := chromedp.Run(ctxw, scrapIt(url, &res, timeout))
	if err != nil {
		log.Fatal(err)
	}
	ch.GetClient().Set(cacheId, res, cache.DefaultExpiration)
	return res
}

func DefaultProxyHandler(ctx *fasthttp.RequestCtx) {
	src, _ := config.GetStaticSrc()
	path := string(ctx.Path())

	fs := &fasthttp.FS{
		Root:               src,
		IndexNames:         []string{config.GetIndexName()},
		GenerateIndexPages: true,
		Compress:           config.IsCompress(),
		AcceptByteRange:    true,
	}
	fsHandler := fs.NewRequestHandler()

	// static proxy
	for postfix, contentType := range StaticPostfix {
		if strings.Contains(strings.ToLower(path), postfix) {
			ctx.SetContentType(contentType)
			fsHandler(ctx)
			return
		}
	}

	ctx.URI().SetPath("/")
	switch string(ctx.Path()) {
	default:
		ctx.SetContentType("text/html")
		fsHandler(ctx)
	}
}

func scrapIt(url string, str *string, t time.Duration) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady("html"),
		chromedp.Sleep(t),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			*str, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}),
	}
}
