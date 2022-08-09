package http01

import (
	"net/http"
	"strings"

	"github.com/darvaza-proxy/darvaza/acme"
)

var (
	_ http.Handler = (*Http01ChallengeHandler)(nil)
)

type Http01ChallengeHandler struct {
	resolver acme.Http01Resolver
	next     http.Handler
}

func (h *Http01ChallengeHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	host := req.URL.Host
	path := req.URL.Path

	h.resolver.AnnounceHost(host)

	token := strings.TrimPrefix(path, "/.well-known/acme-challenge")
	if token == path {
		// invalid prefix
		goto skip
	} else if l := len(token); l == 0 {
		// no token
		goto reject
	} else if token[0] != '/' {
		// invalid prefix
		goto skip
	} else if c := h.resolver.LookupChallenge(host, token[1:]); c == nil {
		// host,token pair not recognised
		goto reject
	} else {
		// host,token pair recognised, proceed
		c.ServeHTTP(rw, req)
		return
	}

skip:
	h.next.ServeHTTP(rw, req)
	return
reject:
	http.NotFound(rw, req)
}

func NewHtt01ChallengeHandler(resolver acme.Http01Resolver) *Http01ChallengeHandler {
	return &Http01ChallengeHandler{
		resolver: resolver,
		next:     NewHttpsRedirectHandler(),
	}
}

func NewHttp01ChallengeMiddleware(resolver acme.Http01Resolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if next == nil {
			next = NewHttpsRedirectHandler()
		}

		return &Http01ChallengeHandler{
			resolver: resolver,
			next:     next,
		}
	}
}
