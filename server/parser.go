package server

import (
	"context"
	"github.com/chromedp/cdproto/dom"
	"github.com/cjp2600/assr/config"
	"github.com/fasthttp/router"
	"github.com/patrickmn/go-cache"
	"github.com/valyala/fasthttp"
	proxy "github.com/yeqown/fasthttp-reverse-proxy"
	"log"
	//"time"

	"github.com/chromedp/chromedp"
	"strings"
)

var (
	staticProxyServer = proxy.NewReverseProxy("localhost:8080")
	inm               *cache.Cache
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (s *Parser) Run() error {
	if !config.IsDebug() {
		proxy.SetProduction()
	}
	inm = cache.New(config.GetCacheDurationTTL(), config.GetCacheDurationTTL())

	r := router.New()
	r.NotFound = DefaultProxyHandler

	for _, route := range config.GetRoutes() {
		r.GET(route, SsrHandler)
	}

	return fasthttp.ListenAndServe(":"+config.GetAppPort(), r.Handler)
}

func SsrHandler(ctx *fasthttp.RequestCtx) {
	var path = string(ctx.Path())
	var cacheId = "cache" + path
	var res string
	html, found := inm.Get(cacheId)
	if found {
		res = html.(string)
	} else {
		var url = config.GetStaticDomain(path)
		var ops []chromedp.ContextOption
		if config.IsDebug() {
			ops = append(ops, chromedp.WithDebugf(log.Printf))
		}
		ctxw, cancel := chromedp.NewContext(context.Background(), ops...)
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
	// proxy static
	for sp, ct := range StaticPostfix {
		if strings.Contains(strings.ToLower(string(ctx.Path())), sp) {
			ctx.SetContentType(ct)
			staticProxyServer.ServeHTTP(ctx)
			return
		}
	}

	ctx.SetContentType("text/html; charset=utf-8")
	staticProxyServer.ServeHTTP(ctx)
}

func scrapIt(url string, str *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady("html"),
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
