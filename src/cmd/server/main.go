package main

import (
	"flag"

	"github.com/TheJubadze/RateLimiter/pkg/app"
)

var configFile = flag.String("config", "/etc/rate-limiter/config.yaml", "Path to configuration file")

func init() {
	flag.Parse()
}

func main() {
	app.StartServer(configFile)
}
