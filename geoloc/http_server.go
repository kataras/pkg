package geoloc

import (
	"regexp"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

// Optionally http server.

var regexIP, _ = regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)

func h(ctx iris.Context, ip string) {
	if !regexIP.MatchString(ip) {
		ctx.NotFound()
		return
	}

	jsonpCallback := ctx.URLParam("callback")
	indent := ""
	if pretty := ctx.URLParamExists("pretty"); pretty {
		indent = " "
	}

	info, ok := Fetch(ip)

	if !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	if jsonpCallback != "" {
		ctx.JSONP(info, context.JSONP{
			Callback: jsonpCallback,
			Indent:   indent,
		})
		return
	}

	ctx.JSON(info, context.JSON{Indent: indent})
}

func setupServer() *iris.Application {
	app := iris.New()

	app.Get("/{ip:string}", func(ctx context.Context) {
		// we could add the regexp on the macro but let's do this here
		ip := ctx.Params().Get("ip")
		h(ctx, ip)
	})

	app.Get("/me", func(ctx context.Context) {
		ip := ctx.RemoteAddr()
		h(ctx, ip)
	})

	return app
}

// Listen starts the http server on the "hostname:port".
func Listen(addr string) error {
	app := setupServer()
	return app.Run(iris.Addr(addr))
}

// Run runs the server using an iris runner (ipv4 address, tls/ssl, automatic ssl and more).
func Run(runner iris.Runner) error {
	app := setupServer()
	return app.Run(runner)
}
