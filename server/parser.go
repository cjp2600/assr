package server

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	ch "github.com/cjp2600/assr/cache"
	"github.com/cjp2600/assr/config"
	"github.com/fasthttp/router"
	"github.com/patrickmn/go-cache"
	"github.com/savsgio/gotils"
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
		if strings.Contains(rt.Path, "{") {
			wg.Done()
			continue
		}
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

	userAgent := "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36"
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.UserAgent(userAgent),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctxw, cancel := chromedp.NewContext(allocCtx, ops...)
	defer cancel()

	chromedp.ListenTarget(ctxw, DisableImageLoad(ctxw))

	err := chromedp.Run(ctxw, scrapIt(url, &res, timeout))
	if err != nil {
		log.Fatal(err)
	}
	ch.GetClient().Set(cacheId, res, cache.DefaultExpiration)
	return res
}

func DisableImageLoad(ctx context.Context) func(event interface{}) {
	return func(event interface{}) {
		switch ev := event.(type) {
		case *fetch.EventRequestPaused:
			go func() {
				c := chromedp.FromContext(ctx)
				ctx := cdp.WithExecutor(ctx, c.Target)

				if ev.ResourceType == network.ResourceTypeImage {
					fetch.FailRequest(ev.RequestID, network.ErrorReasonBlockedByClient).Do(ctx)
				} else {
					fetch.ContinueRequest(ev.RequestID).Do(ctx)
				}
			}()
		}
	}
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

func GetOptionalPaths(path string) []string {
	paths := make([]string, 0)

	start := 0
walk:
	for {
		if start >= len(path) {
			return paths
		}

		c := path[start]
		start++

		if c != '{' {
			continue
		}

		newPath := ""
		questionMarkIndex := -1

		for end, c := range []byte(path[start:]) {
			switch c {
			case '}':
				if questionMarkIndex == -1 {
					continue walk
				}

				end++
				newPath += path[questionMarkIndex+1 : start+end]

				path = path[:questionMarkIndex] + path[questionMarkIndex+1:] // remove '?'
				paths = append(paths, newPath)
				start += end - 1

				continue walk

			case '?':
				questionMarkIndex = start + end
				newPath += path[:questionMarkIndex]

				// include the path without the wildcard
				// -2 due to remove the '/' and '{'
				if !gotils.StringSliceInclude(paths, path[:start-2]) {
					paths = append(paths, path[:start-2])
				}
			}
		}
	}
}
