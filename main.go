package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Proxy struct {
	SubPath string
	URI     *url.URL
}

func main() {
	port := 8080

	root := &cobra.Command{
		Short: "proxyutil is a simple proxy server",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			proxies := make([]Proxy, 0, len(args)-1)

			for _, arg := range args {
				proxy, err := parseProxy(arg)
				if err != nil {
					logrus.Fatalln(err)
				}
				proxies = append(proxies, proxy)
			}

			for i := range proxies {
				proxy := proxies[i]

				logrus.Infof("Proxying: %s -> %s", proxy.SubPath, proxy.URI)
			}

			r := chi.NewRouter()
			r.Use(middleware.Logger)
			for i := range proxies {
				proxy := proxies[i]

				r.Mount(proxy.SubPath, proxyHandler(proxy))
			}

			PORT := fmt.Sprintf(":%d", port)

			logrus.Infof("Listening on %s", PORT)
			err := http.ListenAndServe(PORT, r)
			if err != nil {
				logrus.Fatalln(err)
			}
		},
	}

	root.Flags().IntVarP(&port, "port", "p", port, "port to listen on")
	logrus.Fatal(root.Execute())
}

func proxyHandler(proxy Proxy) http.Handler {
	proxyH := httputil.NewSingleHostReverseProxy(proxy.URI)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyH.ServeHTTP(w, r)
	})
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
