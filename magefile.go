//go:build mage

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/fsnotify/fsnotify"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

const ProjectName = "gwsn"

type TargetOS string

const (
	TargetOS_linux   TargetOS = "linux"
	TargetOS_windows TargetOS = "windows"
	TargetOS_darwin  TargetOS = "darwin"
	TargetOS_freebsd TargetOS = "freebsd"

	TargetOS_unsupported = ""
)

var ErrUnsupportedTargetOS = errors.New("unsupported target OS")

type TargetArch string

const (
	TargetArch_amd64 TargetArch = "amd64"
	TargetArch_arm64 TargetArch = "arm64"
	TargetArch_386   TargetArch = "386"

	TargetArch_unsupported = ""
)

var ErrUnsupportedTargetArch = errors.New("unsupported target arch")

var GlobPattern_entrypoint = "main.go"

var GlobPatterns_assets = []string{
	"internal/**/assets/*.png",
	"internal/**/assets/*.ico",
}

var GlobPatterns_sources = append([]string{
	"internal/**/*.go",
}, GlobPattern_entrypoint)

var GlobPatterns_module = []string{
	"go.mod",
	"go.sum",
}

var Default = BuildHost

func Run(ctx context.Context, targetOS, targetArch string) error {
	mg.Deps(mg.F(Build, targetOS, targetArch))

	resolvedTargetOS, err := resolveTargetOS(targetOS)
	if err != nil {
		return fmt.Errorf("failed to resolve target OS %s: %v", targetOS, err)
	}

	resolvedTargetArch, err := resolveTargetArch(targetArch)
	if err != nil {
		return fmt.Errorf("failed to resolve target arch %s: %v", targetArch, err)
	}

	dst := resolveOutputFile(resolvedTargetOS, resolvedTargetArch)

	cmd := exec.Command(dst)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func RunHost() {
	mg.Deps(mg.F(Run, runtime.GOOS, runtime.GOARCH))
}

func Build(ctx context.Context, targetOS, targetArch string) error {
	resolvedTargetOS, err := resolveTargetOS(targetOS)
	if err != nil {
		return fmt.Errorf("failed to resolve target OS %s: %v", targetOS, err)
	}

	resolvedTargetArch, err := resolveTargetArch(targetArch)
	if err != nil {
		return fmt.Errorf("failed to resolve target arch %s: %v", targetArch, err)
	}

	dst := resolveOutputFile(resolvedTargetOS, resolvedTargetArch)
	newer, err := target.Glob(dst, slices.Concat(GlobPatterns_sources, GlobPatterns_assets, GlobPatterns_module)...)
	if err != nil {
		return fmt.Errorf("failed to check for changes: %v", err)
	}

	if !newer {
		if mg.Verbose() {
			fmt.Println("no changes")
		}
		return nil
	}

	mg.Deps(InstallDeps)

	fmt.Printf("Building %s-%s...\n", resolvedTargetArch, resolvedTargetOS)
	return sh.RunWith(map[string]string{
		"GOOS":   string(resolvedTargetOS),
		"GOARCH": string(resolvedTargetArch),
	}, "go", "build", "-o", dst, GlobPattern_entrypoint)
}

func BuildHost() {
	mg.Deps(mg.F(Build, runtime.GOOS, runtime.GOARCH))
}

func BuildAll() {
	f := func(os TargetOS, arch TargetArch) mg.Fn {
		return mg.F(Build, string(os), string(arch))
	}

	mg.Deps(
		f(TargetOS_linux, TargetArch_amd64),

		f(TargetOS_windows, TargetArch_amd64),
		f(TargetOS_windows, TargetArch_arm64),
	)
}

// Manage your deps, or running package managers.
func InstallDeps() error {
	fmt.Println("Installing dependencies...")
	return sh.Run("go", "mod", "tidy")
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("build")
}

func Watch(ctx context.Context) error {
	fmt.Println("Starting watch mode...")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %v", err)
	}
	defer watcher.Close()

	updateWatchList := func() error {
		newWatchList := make([]string, 0)

		for _, pat := range slices.Concat(GlobPatterns_sources, GlobPatterns_assets, GlobPatterns_module) {
			fs, err := filepath.Glob(pat)
			if err != nil {
				return fmt.Errorf("failed to glob files with pattern %s: %v", pat, err)
			}

			for _, f := range fs {
				abs, err := filepath.Abs(f)
				if err != nil {
					return fmt.Errorf("failed to resolve absolute path for %s: %v", f, err)
				}

				dir := filepath.Dir(abs)
				err = watcher.Add(dir)
				if err != nil {
					return fmt.Errorf("failed to watch directory for changes %s: %v", dir, err)
				}

				if !slices.Contains(newWatchList, dir) {
					newWatchList = append(newWatchList, dir)
				}
			}
		}

		watchList := watcher.WatchList()
		for _, w := range watchList {
			if !slices.Contains(newWatchList, w) {
				err := watcher.Remove(w)

				// ignore ErrNonExistentWatch
				if err != nil && err != fsnotify.ErrNonExistentWatch {
					return fmt.Errorf("failed remove directory from watch list %s: %v", w, err)
				}
			}
		}

		return nil
	}

	var cmd *exec.Cmd = nil
	handleWatchEvent := func() error {
		if cmd != nil {
			cmd.Process.Kill()
			cmd = nil
		}

		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable: %v", err)
		}

		cmd = exec.Command(exe, "Run")
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Start()

		//if err := cmd.Run(); err != nil {
		// Don't return an error here. If the build or run fails,
		// just wait and try again later
		//	fmt.Fprintf(os.Stderr, "process exited with code %d", cmd.ProcessState.ExitCode())
		//} else {
		//	fmt.Println("build completed successfully")
		//}

		if err := updateWatchList(); err != nil {
			return fmt.Errorf("failed to update watch list: %v", err)
		}

		return nil
	}

	if err := handleWatchEvent(); err != nil {
		return fmt.Errorf("error while handling watch event: %v", err)
	}

	for {
		fmt.Println("Waiting for change to trigger rebuild...")

		select {
		case ev, ok := <-watcher.Events:
			if !ok {
				break
			}

			if mg.Verbose() {
				fmt.Printf("received event from watcher (op = %s, file = %s)\n", ev.Op.String(), ev.Name)
			}

			shouldRebuild := false
			for _, pat := range slices.Concat(GlobPatterns_sources, GlobPatterns_assets, GlobPatterns_module) {
				fs, err := filepath.Glob(pat)
				if err != nil {
					return fmt.Errorf("failed to glob files with pattern %s: %v", pat, err)
				}

				if slices.Contains(fs, ev.Name) {
					shouldRebuild = true
					break
				}
			}

			if shouldRebuild {
				fmt.Println("Detected change, triggering rebuild...")
				if err := handleWatchEvent(); err != nil {
					return fmt.Errorf("error while handling watch event: %v", err)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				break
			}

			return fmt.Errorf("error while watching for file change: %v", err)

		case <-ctx.Done():
			return nil
		}
	}
}

func resolveTargetOS(s string) (TargetOS, error) {
	switch s {
	case "linux":
		return TargetOS_linux, nil
	case "win", "windows":
		return TargetOS_windows, nil
	case "darwin", "mac", "macos":
		return TargetOS_darwin, nil
	case "freebsd":
		return TargetOS_freebsd, nil
	}

	return TargetOS_unsupported, ErrUnsupportedTargetOS
}

func resolveTargetArch(s string) (TargetArch, error) {
	switch s {
	case "amd64", "x86_64":
		return TargetArch_amd64, nil
	case "arm64", "apple":
		return TargetArch_arm64, nil
	case "i386", "386":
		return TargetArch_386, nil
	}

	return TargetArch_unsupported, ErrUnsupportedTargetArch
}

func resolveTargetFileExt(targetOs TargetOS) string {
	switch targetOs {
	case TargetOS_windows:
		return ".exe"
	default:
		return ""
	}
}

func resolveOutputFile(targetOs TargetOS, targetArch TargetArch) string {
	ext := resolveTargetFileExt(targetOs)
	return fmt.Sprintf("build/%s/%s-%s/%s%s", ProjectName, targetArch, targetOs, ProjectName, ext)
}
