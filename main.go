package main

import (
	"runtime"
	"strconv"

	"github.com/backupflex/core/internal"
	"github.com/go-core-fx/healthfx"
	"github.com/samber/lo"
)

//go:generate swag init --parseDependency --outputTypes go -g ./main.go -o ./internal/server/docs

//nolint:gochecknoglobals // build metadata
var (
	appVersion   = "dev"
	appReleaseID = "0"
	appBuildDate = "unknown"
	appGitCommit = "unknown"
	appGoVersion = runtime.Version()
)

func main() {
	internal.Run(healthfx.Version{
		Version:   appVersion,
		ReleaseID: lo.Must1(strconv.Atoi(appReleaseID)),
		BuildDate: appBuildDate,
		GitCommit: appGitCommit,
		GoVersion: appGoVersion,
	})
}
