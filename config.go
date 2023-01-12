package main

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"strings"
)

var defaultConfigFilename = ".proxies"

func Parse(r io.Reader) ([]Proxy, error) {
	br := bufio.NewReader(r)

	var proxies []Proxy
	var readErr error

	for readErr != io.EOF {
		s, readErr := br.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return nil, readErr
		}

		s = strings.TrimSpace(s)
		if s == "" {
			return proxies, nil
		}

		p, err := parseProxy(s)
		if err != nil {
			return nil, err
		}

		proxies = append(proxies, p)
	}

	return proxies, nil
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
