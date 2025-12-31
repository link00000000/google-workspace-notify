assets := $(wildcard internal/**/*.png) $(wildcard internal/**/*.ico)
entrypoint := main.go
sources := $(entrypoint) $(wildcard internal/**/*.go)
module := go.mod go.sum

.PHONY : default
default : build/gwsn/amd64-windows/gwsn.exe

.PHONY : run
run : build/gwsn/amd64-windows/gwsn.exe
	$<

.PHONY : clean
clean :
	@rm -r -f build

# Linux builds
build/gwsn/amd64-linux/gwsn : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=amd64 go build -o $@ $(entrypoint)

build/gwsn/arm64-linux/gwsn : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=arm64 go build -o $@ $(entrypoint)

build/gwsn/386-linux/gwsn : $(sources) $(assets) $(module)
	GOOS=linux GOARCH=386 go build -o $@ $(entrypoint)

# macOS builds
build/gwsn/amd64-darwin/gwsn : $(sources) $(assets) $(module)
	GOOS=darwin GOARCH=amd64 go build -o $@ $(entrypoint)

build/gwsn/arm64-darwin/gwsn : $(sources) $(assets) $(module)
	GOOS=darwin GOARCH=arm64 go build -o $@ $(entrypoint)

# FreeBSD builds
build/gwsn/amd64-freebsd/gwsn : $(sources) $(assets) $(module)
	GOOS=freebsd GOARCH=amd64 go build -o $@ $(entrypoint)

# Windows builds
build/gwsn/amd64-windows/gwsn.exe : $(sources) $(assets) $(module)
	GOOS=windows GOARCH=amd64 go build -o $@ $(entrypoint)

build/gwsn/386-windows/gwsn.exe : $(sources) $(assets) $(module)
	GOOS=windows GOARCH=386 go build -o $@ $(entrypoint)
