package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

func GetStaticDomain(path string) string {
	return "http://localhost:" + GetStaticPort() + path
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

func GetRoutes() []string {
	if !IsEnableSSR() {
		return []string{}
	}
	routes := viper.GetStringSlice("ssr.routes")
	return routes
}

func GetIndexName() string {
	indexName := viper.GetString("app.index")
	if len(indexName) > 0 {
		return indexName
	}
	return "index.html"
}
