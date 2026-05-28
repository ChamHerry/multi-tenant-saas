package cmd

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var spaStaticRootCandidates = []string{
	"web/dist",
	filepath.Join("resource", "public"),
}

func configureStaticFileService(ctx context.Context, s *ghttp.Server) {
	root, ok := findSPAStaticRoot()
	if !ok {
		g.Log().Debug(ctx, "SPA static root not found; skip frontend static service")
		return
	}

	s.SetIndexFolder(false)
	s.SetServerRoot(root)
	configureSPAFallback(root, s)
	g.Log().Infof(ctx, "SPA static service enabled root=%s", root)
}

func configureSPAFallback(root string, s *ghttp.Server) {
	indexPath := filepath.Join(root, "index.html")
	s.BindStatusHandler(http.StatusNotFound, func(r *ghttp.Request) {
		if !shouldServeSPAFallback(r.Method, r.URL.Path) {
			return
		}
		if _, err := os.Stat(indexPath); err != nil {
			return
		}
		r.Response.ClearBuffer()
		r.Response.Status = http.StatusOK
		r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		r.Response.ServeFile(indexPath)
	})
}

func findSPAStaticRoot() (string, bool) {
	for _, candidate := range spaStaticRootCandidates {
		if root, ok := validSPAStaticRoot(candidate); ok {
			return root, true
		}
	}
	if executable, err := os.Executable(); err == nil {
		base := filepath.Dir(executable)
		for _, candidate := range spaStaticRootCandidates {
			if root, ok := validSPAStaticRoot(filepath.Join(base, candidate)); ok {
				return root, true
			}
		}
	}
	return "", false
}

func validSPAStaticRoot(candidate string) (string, bool) {
	root, err := filepath.Abs(candidate)
	if err != nil {
		return "", false
	}
	info, err := os.Stat(filepath.Join(root, "index.html"))
	if err != nil || info.IsDir() {
		return "", false
	}
	return root, true
}

func shouldServeSPAFallback(method, requestPath string) bool {
	if method != http.MethodGet && method != http.MethodHead {
		return false
	}
	if requestPath == "" {
		requestPath = "/"
	}
	for _, skipped := range []string{"/api", "/healthz", "/readyz", "/swagger", "/api.json"} {
		if requestPath == skipped || strings.HasPrefix(requestPath, skipped+"/") {
			return false
		}
	}
	if ext := filepath.Ext(requestPath); ext != "" {
		return false
	}
	return true
}
