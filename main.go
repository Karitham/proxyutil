package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	stdlog "log"
)

type Proxy struct {
	SubPath string
	URI     *url.URL
}

var log = logrus.New()

func main() {
	port := 8080
	level := uint(2)

	app := &cli.App{
		Name:  "proxyutil",
		Usage: "proxyutil is a simple proxy server",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				EnvVars:     []string{"PROXYUTIL_PORT"},
				Value:       8080,
				Usage:       "port to listen on",
				Destination: &port,
			},
			&cli.UintFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				EnvVars:     []string{"PROXYUTIL_VERBOSE"},
				Value:       4,
				Usage:       "enable verbose logging. 0 = panic, 1 = fatal, 2 = error, 3 = warn, 4 = info, 5 = debug, 6 = trace",
				Destination: &level,
			},
		},

		Before: func(ctx *cli.Context) error {
			log.SetLevel(logrus.Level(level))
			return nil
		},

		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return fmt.Errorf("no proxy specified")
			}

			args := c.Args().Slice()
			proxies := make([]Proxy, 0, len(args))

			for _, arg := range args {
				proxy, err := parseProxy(arg)
				if err != nil {
					log.Fatalln(err)
				}
				proxies = append(proxies, proxy)
			}

			for i := range proxies {
				proxy := proxies[i]

				log.Infof("Proxying: %s -> %s", proxy.SubPath, proxy.URI)
			}

			r := chi.NewRouter()
			r.Use(reqLogger(log))

			for i := range proxies {
				proxy := httputil.NewSingleHostReverseProxy(proxies[i].URI)

				// logger
				lg := &stdlog.Logger{}
				lg.SetOutput(log.WithField("path", proxies[i].SubPath).WriterLevel(logrus.WarnLevel))
				proxy.ErrorLog = lg

				r.Mount(proxies[i].SubPath, proxy)
			}

			addr := fmt.Sprintf(":%d", port)

			log.Infof("Listening on %s", addr)
			return http.ListenAndServe(addr, r)
		},
	}

	log.Fatal(app.Run(os.Args))
}

func parseProxy(arg string) (Proxy, error) {
	subP, uri, ok := strings.Cut(arg, ":")
	if !ok {
		return Proxy{}, fmt.Errorf("invalid proxy: %s", arg)
	}

	u, err := url.Parse(uri)
	if err != nil {
		return Proxy{}, fmt.Errorf("invalid uri: %s", uri)
	}

	return Proxy{
		SubPath: subP,
		URI:     u,
	}, nil
}

// reqLogger returns a request logging middleware
func reqLogger(logger logrus.FieldLogger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			reqID := middleware.GetReqID(r.Context())
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					remoteIP = r.RemoteAddr
				}
				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}
				fields := logrus.Fields{
					"status_code": ww.Status(),
					"bytes":       ww.BytesWritten(),
					"duration":    time.Since(t1).String(),
					"remote_ip":   remoteIP,
					"method":      r.Method,
				}
				if len(reqID) > 0 {
					fields["request_id"] = reqID
				}
				logger.WithFields(fields).Debugf("%s://%s%s", scheme, r.Host, r.RequestURI)
			}()

			h.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
