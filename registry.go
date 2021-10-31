package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type registry interface {
	add(name string, deps []string) error
	remove(name string) error
	list() string
}

type onePackage struct {
	name       string
	dependsOn  []string
	requiredBy []string
}

func (pkg onePackage) String() string {
	return fmt.Sprintf("package %s with deps %q and required by %q", pkg.name, pkg.dependsOn, pkg.requiredBy)
}

type inMemoryStore struct {
	sync.RWMutex
	packages map[string]onePackage
}

func newInMemoryStore() *inMemoryStore {
	return &inMemoryStore{
		packages: make(map[string]onePackage),
	}
}

func (store *inMemoryStore) add(name string, deps []string) error {
	store.Lock()
	defer store.Unlock()

	if pkg, exists := store.packages[name]; exists {
		return fmt.Errorf("package already exists: %s", pkg.String())
	}
	var validDeps []string
	// tell dependencies that this package is depending on them
	for _, dep := range deps {
		if pkg, exists := store.packages[dep]; exists {
			pkg.requiredBy = append(pkg.requiredBy, name)
			store.packages[dep] = pkg
			validDeps = append(validDeps, dep)
		}
	}
	// add package to the registry
	store.packages[name] = onePackage{
		name:      name,
		dependsOn: validDeps,
	}
	return nil
}

func (store *inMemoryStore) remove(name string) error {
	store.Lock()
	defer store.Unlock()

	toRemove, exists := store.packages[name]
	if !exists {
		return fmt.Errorf("package not exists: %s", name)
	}
	if len(toRemove.requiredBy) > 0 {
		return fmt.Errorf("package %s cannot be removed, it's required by %q", name, toRemove.requiredBy)
	}
	// tell dependencies that this package is no longer depending on them
	for _, dep := range toRemove.dependsOn {
		store.removeRequiredBy(dep, name)
	}
	// remove package from registry
	delete(store.packages, name)
	return nil
}

func (store *inMemoryStore) removeRequiredBy(pkgName, notRequiredAnymore string) {
	if pkg, exists := store.packages[pkgName]; exists {
		var stillRequiredBy []string
		for _, requiredByPkgName := range pkg.requiredBy {
			if requiredByPkgName != notRequiredAnymore {
				stillRequiredBy = append(stillRequiredBy, requiredByPkgName)
			}
		}
		pkg.requiredBy = stillRequiredBy
		store.packages[pkgName] = pkg
	}
}

func (store *inMemoryStore) list() string {
	store.RLock()
	defer store.RUnlock()

	output := "Packages and Dependencies\n"
	if len(store.packages) == 0 {
		output += "- No packages found"
	} else {
		pkgs := store.packages
		keys := make([]string, 0, len(pkgs))
		for key := range pkgs {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			output += store.listOnePackage(key, 0)
		}
	}
	return strings.TrimRight(output, "\n")
}

func (store *inMemoryStore) listOnePackage(name string, level int) string {
	pkg, exists := store.packages[name]
	if !exists {
		return ""
	}
	indentation := strings.Repeat(" ", level*4)
	output := indentation + fmt.Sprintf("- %s\n", pkg.name)
	deps := pkg.dependsOn
	sort.Strings(deps)
	for _, dep := range deps {
		if tree := store.listOnePackage(dep, level+1); tree != "" {
			output += strings.TrimRight(tree, "\n") + "\n"
		}
	}
	return output
}
