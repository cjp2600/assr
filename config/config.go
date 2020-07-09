package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
	"time"
)

type Routes struct {
	Path    string
	Timeout time.Duration
}

func GetStaticDomain(path string) string {
	return "http://localhost:" + GetStaticPort() + path
}

func GetWaitTimeout() time.Duration {
	ttl := viper.GetDuration("ssr.sleep")
	return ttl * time.Millisecond
}

func GetAppDomain() string {
	return "http://localhost:" + GetAppPort()
}

func GetAppPort() string {
	port := viper.GetString("app.port")
	if len(port) == 0 {
		return "8080"
	}
	return port
}

func GetStaticSrc() (string, error) {
	src := viper.GetString("static.src")
	if len(src) == 0 {
		return "", fmt.Errorf("static folder not found (static.src config file)")
	}
	return src, nil
}

func GetStaticPort() string {
	port := viper.GetString("static.port")
	if len(port) == 0 {
		return "3000"
	}
	return port
}

func GetCacheDurationTTL() time.Duration {
	ttl := viper.GetDuration("cache.ttl")
	return ttl * time.Second
}

func GetProjectName() string {
	return viper.GetString("project")
}

func IsDebug() bool {
	return viper.GetBool("app.debug")
}

func IsCompress() bool {
	return viper.GetBool("static.compress")
}

func IsEnableSSR() bool {
	return viper.GetBool("ssr.enable")
}

func GetRouteTimeout(path []byte) time.Duration {
	for _, item := range GetRoutes() {
		if strings.EqualFold(item.Path, string(path)) {
			return item.Timeout
		}
	}
	return 0
}

func GetRoutes() []Routes {
	var res []Routes
	if !IsEnableSSR() {
		return nil
	}
	routes := viper.Get("ssr.routes")
	if it, ok := routes.([]interface{}); ok {
		for _, item := range it {
			var path string
			var timeout int
			for k, v := range item.(map[interface{}]interface{}) {
				if k.(string) == "path" {
					path = v.(string)
				}
				timeout = 2500
				if k.(string) == "timeout" {
					timeout = v.(int)
				}
			}
			res = append(res, Routes{
				Path:    path,
				Timeout: time.Duration(timeout) * time.Millisecond,
			})
		}
	}

	return res
}

func GetIndexName() string {
	indexName := viper.GetString("app.index")
	if len(indexName) > 0 {
		return indexName
	}
	return "index.html"
}
