package main

const service = "identity"


import (
    cfgpkg "insta-backend/pkg/config"
    logpkg "insta-backend/pkg/logger"
)


func main() {
	cfg, err := cfgpkg.Load(service)
}