package main

import (
	"fmt"
	"strings"
	"sync"
)

type onePackage struct {
	name       string
	dependsOn  []string
	requiredBy []string
}

func (pkg onePackage) String() string {
	return fmt.Sprintf("package %s with deps %q and required by %q", pkg.name, pkg.dependsOn, pkg.requiredBy)
}

type registry struct {
	sync.RWMutex
	packages map[string]onePackage
}

func newRegistry() *registry {
	return &registry{
		packages: make(map[string]onePackage),
	}
}

func (reg *registry) add(name string, deps []string) error {
	reg.Lock()
	defer reg.Unlock()

	if pkg, exists := reg.packages[name]; exists {
		return fmt.Errorf("package already exists: %s", pkg.String())
	}
	// add package to the registry
	reg.packages[name] = onePackage{
		name:      name,
		dependsOn: deps,
	}
	// tell dependencies that this package is depending on them
	for _, dep := range deps {
		if pkg, exists := reg.packages[dep]; exists {
			pkg.requiredBy = append(pkg.requiredBy, name)
			reg.packages[dep] = pkg
		}
	}
	return nil
}

func (reg *registry) remove(name string) error {
	reg.Lock()
	defer reg.Unlock()

	toRemove, exists := reg.packages[name]
	if !exists {
		return fmt.Errorf("package not exists: %s", name)
	}
	if len(toRemove.requiredBy) > 0 {
		return fmt.Errorf("package %s cannot be removed, it's is required by %q", name, toRemove.requiredBy)
	}
	// tell dependencies that this package is no longer depending on them
	for _, dep := range toRemove.dependsOn {
		reg.removeRequiredBy(dep, name)
	}
	// remove package from registry
	delete(reg.packages, name)
	return nil
}

func (reg *registry) removeRequiredBy(pkgName, requiredBy string) {
	if pkg, exists := reg.packages[pkgName]; exists {
		var stillRequiredBy []string
		for _, requiredByPkgName := range pkg.requiredBy {
			if requiredByPkgName != requiredBy {
				stillRequiredBy = append(stillRequiredBy, requiredByPkgName)
			}
		}
		pkg.requiredBy = stillRequiredBy
		reg.packages[pkgName] = pkg
	}
}

func (reg *registry) list() string {
	reg.RLock()
	defer reg.RUnlock()

	output := "Packages and Dependencies\n"
	if len(reg.packages) == 0 {
		output += "- No packages found"
	} else {
		for _, pkg := range reg.packages {
			output += reg.listOnePackage(pkg.name, 0)
		}
	}
	return output
}

func (reg *registry) listOnePackage(name string, level int) string {
	pkg, exists := reg.packages[name]
	if !exists {
		return ""
	}
	indentation := strings.Repeat(" ", level*4)
	output := indentation + fmt.Sprintf("- %s\n", pkg.name)
	for _, dep := range pkg.dependsOn {
		if tree := reg.listOnePackage(dep, level+1); tree != "" {
			output += strings.TrimRight(tree, "\n") + "\n"
		}
	}
	return output
}
