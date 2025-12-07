assets := $(wildcard cmd/**/assets/*.png) $(wildcard cmd/**/assets/*.ico)
sources := $(wildcard src/**/*.go) $(wildcard cmd/**/*.go)
module := go.mod go.sum

default : build/google-workspace-notify/x86_64-windows/google-workspace-notify.exe

# Linux builds
build/google-workspace-notify/x86_64-linux/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=amd64 go build -o $@ ./cmd/google-workspace-notify

build/google-workspace-notify/arm64-linux/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=arm64 go build -o $@ ./cmd/google-workspace-notify

build/google-workspace-notify/386-linux/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=386 go build -o $@ ./cmd/google-workspace-notify

# macOS builds
build/google-workspace-notify/x86_64-darwin/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=darwin GOARCH=amd64 go build -o $@ ./cmd/google-workspace-notify

build/google-workspace-notify/arm64-darwin/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=darwin GOARCH=arm64 go build -o $@ ./cmd/google-workspace-notify

# FreeBSD builds
build/google-workspace-notify/x86_64-freebsd/google-workspace-notify : $(sources) $(assets) $(module)
	GOOS=freebsd GOARCH=amd64 go build -o $@ ./cmd/google-workspace-notify

# Windows builds
build/google-workspace-notify/x86_64-windows/google-workspace-notify.exe : $(sources) $(assets) $(module)
	GOOS=windows GOARCH=amd64 go build -o $@ ./cmd/google-workspace-notify

build/google-workspace-notify/386-windows/google-workspace-notify.exe : $(sources) $(assets) $(module)
	GOOS=windows GOARCH=386 go build -o $@ ./cmd/google-workspace-notify
