package module

import (
	"context"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
	"strings"
)

type Dependency struct {
	Module   *packages.Module
	Packages []*packages.Package
}

type DependencyResolver interface {
	Resolve(ctx context.Context, patterns ...string) ([]Dependency, error)
}

func NewDependencyResolver() DependencyResolver {
	return &dependencyResolver{}
}

type dependencyResolver struct {
}

func (r *dependencyResolver) Resolve(ctx context.Context, patterns ...string) ([]Dependency, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedImports | packages.NeedDeps | packages.NeedFiles | packages.NeedName | packages.NeedModule,
	}
	rootPKGs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, err
	}
	var errs []error
	dependenciesByModule := make(map[string]Dependency)
	packages.Visit(rootPKGs, func(p *packages.Package) bool {
		if len(p.Errors) > 0 {
			for _, err := range p.Errors {
				errs = append(errs, err)
			}
			return false
		}
		if isStdLib(p) {
			return false
		}
		if p.Module == nil {
			errs = append(errs, errors.Errorf("no module for %s", p.PkgPath))
			return false
		}
		dependency, ok := dependenciesByModule[p.Module.Path]
		if !ok {
			dependency = Dependency{
				Module: p.Module,
			}
		}
		dependency.Packages = append(dependency.Packages, p)
		dependenciesByModule[p.Module.Path] = dependency
		return true
	}, nil)
	if len(errs) != 0 {
		var messages []string
		for _, err := range errs {
			messages = append(messages, err.Error())
		}
		return nil, errors.New(strings.Join(messages, ""))
	}

	var dependencies []Dependency
	for _, dependency := range dependenciesByModule {
		dependencies = append(dependencies, dependency)
	}
	return dependencies, nil
}
