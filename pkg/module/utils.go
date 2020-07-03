package module

import (
	"go/build"
	"golang.org/x/tools/go/packages"
	"strings"
)

func isStdLib(pkg *packages.Package) bool {
	if pkg.PkgPath == "unsafe" {
		return true
	}
	if len(pkg.GoFiles) == 0 {
		return false
	}
	return strings.HasPrefix(pkg.GoFiles[0], build.Default.GOROOT)
}
