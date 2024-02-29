package sense

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/creamsensation/sense/config"
	"github.com/creamsensation/sense/internal/constant/contentType"
	"github.com/creamsensation/sense/internal/constant/header"
	"github.com/creamsensation/sense/internal/constant/model"
)

func isRequestMultipart(req *http.Request) bool {
	return strings.Contains(req.Header.Get(header.ContentType), contentType.MultipartForm)
}

func getFileSuffixFromName(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func formatPath(path string) string {
	if len(path) == 0 {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if path != "/" {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

func createRoutePattern(method, pathPrefix, path string) string {
	path = pathPrefix + formatPath(path)
	if !strings.Contains(path, "...") {
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		path += "{$}"
	}
	if len(method) > 0 {
		path = method + " " + path
	}
	return path
}

func wrapError(err error) ([]byte, error) {
	return json.Marshal(model.Error{Error: err.Error()})
}

func wrapResult(v any) ([]byte, error) {
	return json.Marshal(model.Json{Result: v})
}

func assertStringToType[T Assert](v string) T {
	result := *new(T)
	switch any(result).(type) {
	case string:
		result = any(escapeString(v)).(T)
	case bool:
		result = any(v == "true").(T)
	case int:
		res, err := strconv.Atoi(v)
		if err == nil {
			result = any(res).(T)
		}
	case float32:
		res, err := strconv.ParseFloat(v, 32)
		if err == nil {
			result = any(float32(res)).(T)
		}
	case float64:
		res, err := strconv.ParseFloat(v, 64)
		if err == nil {
			result = any(res).(T)
		}
	}
	return result
}

func escapeString(value string) string {
	replacer := strings.NewReplacer("<", "&lt;", ">", "&gt;", "'", "", "\"", "", "`", "")
	value = replacer.Replace(value)
	return value
}

func findFirewallsWithPath(path string, firewalls []config.Firewall) []config.Firewall {
	result := make([]config.Firewall, 0)
	for _, firewall := range firewalls {
		if !firewall.Enabled {
			continue
		}
		match := false
		for _, pattern := range firewall.Patterns {
			if regexp.MustCompile(pattern).MatchString(path) {
				match = true
			}
		}
		if match {
			result = append(result, firewall)
		}
	}
	return result
}
