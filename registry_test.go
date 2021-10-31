package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnePackageString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		given onePackage
		want  string
	}{
		{
			name:  "only has name",
			given: onePackage{name: "AAA"},
			want:  "package AAA with deps [] and required by []",
		},
		{
			name: "has name and dependsOn",
			given: onePackage{
				name:      "AAA",
				dependsOn: []string{"BBB", "CCC"},
			},
			want: `package AAA with deps ["BBB" "CCC"] and required by []`,
		},
		{
			name: "has name, dependsOn and requiredBy",
			given: onePackage{
				name:       "AAA",
				dependsOn:  []string{"BBB", "CCC"},
				requiredBy: []string{"DDD", "FFF"},
			},
			want: `package AAA with deps ["BBB" "CCC"] and required by ["DDD" "FFF"]`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.given.String())
		})
	}
}

func TestInMemoryStoreAdd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		givenPkgs map[string]onePackage
		givenName string
		givenDeps []string
		wantError error
		wantPkgs  map[string]onePackage
	}{
		{
			name: "package already exists",
			givenPkgs: map[string]onePackage{
				"AAA": {name: "AAA"},
			},
			givenName: "AAA",
			givenDeps: []string{},
			wantError: errors.New("package already exists: package AAA with deps [] and required by []"),
			wantPkgs: map[string]onePackage{
				"AAA": {name: "AAA"},
			},
		},
		{
			name: "add a package without deps",
			givenPkgs: map[string]onePackage{
				"AAA": {name: "AAA"},
			},
			givenName: "BBB",
			givenDeps: []string{},
			wantError: nil,
			wantPkgs: map[string]onePackage{
				"AAA": {name: "AAA"},
				"BBB": {name: "BBB"},
			},
		},
		{
			name: "add a package with nonexistent deps",
			givenPkgs: map[string]onePackage{
				"AAA": {name: "AAA"},
			},
			givenName: "BBB",
			givenDeps: []string{"CCC"},
			wantError: nil,
			wantPkgs: map[string]onePackage{
				"AAA": {name: "AAA"},
				"BBB": {name: "BBB"},
			},
		},
		{
			name: "happy path",
			givenPkgs: map[string]onePackage{
				"AAA": {name: "AAA"},
			},
			givenName: "BBB",
			givenDeps: []string{"AAA"},
			wantError: nil,
			wantPkgs: map[string]onePackage{
				"AAA": {name: "AAA", requiredBy: []string{"BBB"}},
				"BBB": {name: "BBB", dependsOn: []string{"AAA"}},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := newInMemoryStore()
			store.packages = tc.givenPkgs

			err := store.add(tc.givenName, tc.givenDeps)
			if tc.wantError != nil {
				assert.EqualError(t, err, tc.wantError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantPkgs, store.packages)
			}
		})
	}
}

func TestInMemoryStoreRemove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		givenPkgs map[string]onePackage
		givenName string
		wantError error
		wantPkgs  map[string]onePackage
	}{
		{
			name: "package not exists",
			givenPkgs: map[string]onePackage{
				"AAA": {name: "AAA"},
			},
			givenName: "BBB",
			wantError: errors.New("package not exists: BBB"),
		},
		{
			name: "package cannot be removed when required by others",
			givenPkgs: map[string]onePackage{
				"AAA": {name: "AAA", requiredBy: []string{"BBB"}},
				"BBB": {name: "BBB", dependsOn: []string{"AAA"}},
			},
			givenName: "AAA",
			wantError: errors.New(`package AAA cannot be removed, it's required by ["BBB"]`),
		},
		{
			name: "happy path",
			givenPkgs: map[string]onePackage{
				"AAA": {name: "AAA", requiredBy: []string{"BBB", "CCC"}},
				"BBB": {name: "BBB", dependsOn: []string{"AAA"}},
				"CCC": {name: "CCC", dependsOn: []string{"AAA"}},
			},
			givenName: "CCC",
			wantError: nil,
			wantPkgs: map[string]onePackage{
				"AAA": {name: "AAA", requiredBy: []string{"BBB"}},
				"BBB": {name: "BBB", dependsOn: []string{"AAA"}},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := newInMemoryStore()
			store.packages = tc.givenPkgs

			err := store.remove(tc.givenName)
			if tc.wantError != nil {
				assert.EqualError(t, err, tc.wantError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantPkgs, store.packages)
			}
		})
	}
}

func TestInMemoryStoreList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		given map[string]onePackage
		want  string
	}{
		{
			name:  "no packages",
			given: map[string]onePackage{},
			want: "Packages and Dependencies\n" +
				"- No packages found",
		},
		{
			name: "one package no deps",
			given: map[string]onePackage{
				"AAA": {name: "AAA"},
			},
			want: "Packages and Dependencies\n" +
				"- AAA",
		},
		{
			name: "two packages and one depends another",
			given: map[string]onePackage{
				"AAA": {name: "AAA", requiredBy: []string{"BBB"}},
				"BBB": {name: "BBB", dependsOn: []string{"AAA"}},
			},
			want: "Packages and Dependencies\n" +
				"- AAA\n" +
				"- BBB\n" +
				"    - AAA",
		},
		{
			name: "more packages and dependencies",
			given: map[string]onePackage{
				"AAA": {name: "AAA", requiredBy: []string{"BBB", "DDD"}},
				"BBB": {name: "BBB", dependsOn: []string{"AAA"}, requiredBy: []string{"CCC", "DDD"}},
				"CCC": {name: "CCC", dependsOn: []string{"BBB"}},
				"DDD": {name: "DDD", dependsOn: []string{"AAA", "BBB"}},
				"EEE": {name: "EEE", dependsOn: []string{"DDD"}},
			},
			want: "Packages and Dependencies\n" +
				"- AAA\n" +
				"- BBB\n" +
				"    - AAA\n" +
				"- CCC\n" +
				"    - BBB\n" +
				"        - AAA\n" +
				"- DDD\n" +
				"    - AAA\n" +
				"    - BBB\n" +
				"        - AAA\n" +
				"- EEE\n" +
				"    - DDD\n" +
				"        - AAA\n" +
				"        - BBB\n" +
				"            - AAA",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := newInMemoryStore()
			store.packages = tc.given

			assert.Equal(t, tc.want, store.list())
		})
	}
}
