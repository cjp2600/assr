package server

import (
	"context"
	"github.com/chromedp/cdproto/dom"
	"github.com/cjp2600/assr/config"
	"github.com/fasthttp/router"
	"github.com/patrickmn/go-cache"
	"github.com/valyala/fasthttp"
	"log"
	//"time"

	"github.com/chromedp/chromedp"
	"strings"
)

var (
	inm *cache.Cache
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (s *Parser) Run(dir string) error {
	inm = cache.New(config.GetCacheDurationTTL(), config.GetCacheDurationTTL())

	r := router.New()
	r.NotFound = DefaultProxyHandler

	for _, route := range config.GetRoutes() {
		r.GET(route, SsrHandler)
	}

	// 		ReadBufferSize: 100 * 1024 * 1024 * 1024,
	serv := fasthttp.Server{
		Handler:        r.Handler,
		ReadBufferSize: 100 * 1024,
	}

	return serv.ListenAndServe(":" + config.GetAppPort())
}

func SsrHandler(ctx *fasthttp.RequestCtx) {
	var (
		path    = string(ctx.Path())
		cacheId = "cache" + path
		res     string
	)

	html, found := inm.Get(cacheId)
	if found {
		res = html.(string)
	} else {
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

		err := chromedp.Run(ctxw, scrapIt(url, &res))
		if err != nil {
			log.Fatal(err)
		}
		inm.Set(cacheId, res, cache.DefaultExpiration)
	}
	ctx.SetContentType("text/html; charset=utf-8")
	ctx.Response.SetBody([]byte(res))
	ctx.SetStatusCode(fasthttp.StatusOK)
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

func scrapIt(url string, str *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady("html"),
		chromedp.Sleep(config.GetWaitTimeout()),
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
