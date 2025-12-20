//go:build mage

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
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

func BuildHost() {
	mg.Deps(mg.F(Build, runtime.GOOS, runtime.GOARCH))
}

func All() {
	f := func(os TargetOS, arch TargetArch) mg.Fn {
		return mg.F(Build, string(os), string(arch))
	}

	mg.Deps(
		f(TargetOS_linux, TargetArch_amd64),

		f(TargetOS_windows, TargetArch_amd64),
		f(TargetOS_windows, TargetArch_arm64),
	)
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
		fmt.Println("no changes")
		return nil
	}

	mg.Deps(InstallDeps)

	fmt.Printf("Building %s-%s...\n", resolvedTargetArch, resolvedTargetOS)
	return sh.RunWith(map[string]string{
		"GOOS":   string(resolvedTargetOS),
		"GOARCH": string(resolvedTargetArch),
	}, "go", "build", "-o", dst, GlobPattern_entrypoint)
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

	handleWatchEvent := func() error {
		mg.Deps(BuildHost)

		if err := updateWatchList(); err != nil {
			return fmt.Errorf("error while updating watch list: %v", err)
		}

		return nil
	}

	if err := handleWatchEvent(); err != nil {
		return fmt.Errorf("error while handling watch event: %v", err)
	}

	fmt.Println("Waiting for change to trigger rebuild...")

	for {
		select {
		case _, ok := <-watcher.Events:
			if !ok {
				break
			}

			fmt.Println("Detected change, triggering rebuild...")
			if err := handleWatchEvent(); err != nil {
				return fmt.Errorf("error while handling watch event: %v", err)
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
