assets := $(wildcard internal/**/assets/*.png) $(wildcard internal/**/assets/*.ico)
entrypoint := main.go
sources := $(entrypoint) $(wildcard internal/**/*.go)
module := go.mod go.sum

default : build/gwsn/x86_64-windows/gwsn.exe

# Linux builds
build/gwsn/x86_64-linux/gwsn : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=amd64 go build -o $@ $(entrypoint)

build/gwsn/arm64-linux/gwsn : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=arm64 go build -o $@ $(entrypoint)

build/gwsn/386-linux/gwsn : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=386 go build -o $@ $(entrypoint)

# macOS builds
build/gwsn/x86_64-darwin/gwsn : $(sources) $(assets) $(module)
	GOOS=darwin GOARCH=amd64 go build -o $@ $(entrypoint)

build/gwsn/arm64-darwin/gwsn : $(sources) $(assets) $(module)
	GOOS=darwin GOARCH=arm64 go build -o $@ $(entrypoint)

# FreeBSD builds
build/gwsn/x86_64-freebsd/gwsn : $(sources) $(assets) $(module)
	GOOS=freebsd GOARCH=amd64 go build -o $@ $(entrypoint)

# Windows builds
build/gwsn/x86_64-windows/gwsn.exe : $(sources) $(assets) $(module)
	GOOS=windows GOARCH=amd64 go build -o $@ $(entrypoint)

build/gwsn/386-windows/gwsn.exe : $(sources) $(assets) $(module)
	GOOS=windows GOARCH=386 go build -o $@ $(entrypoint)
