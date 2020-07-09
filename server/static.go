package server

import (
	"github.com/cjp2600/assr/config"
	"github.com/valyala/fasthttp"
	"strings"
)

type Static struct {
}

func NewStatic() *Static {
	return &Static{}
}

func (s *Static) Run(dir string) error {
	fs := &fasthttp.FS{
		Root:               dir,
		IndexNames:         []string{config.GetIndexName()},
		GenerateIndexPages: true,
		Compress:           config.IsCompress(),
		AcceptByteRange:    true,
	}
	fsHandler := fs.NewRequestHandler()

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.SetUserValue("current", string(ctx.Path()))
		path := string(ctx.Path())

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

	// 		ReadBufferSize: 100 * 1024 * 1024 * 1024,
	serv := fasthttp.Server{
		Handler: requestHandler,
	}
	return serv.ListenAndServe(":" + config.GetStaticPort())
}
