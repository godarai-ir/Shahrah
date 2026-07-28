package providervalidation

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"testing"
)

func TestRegistryExactlyMatchesRuntimeReadyConnectionSchemas(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve test file path")
	}

	migrationsRoot := filepath.Clean(filepath.Join(
		filepath.Dir(currentFile),
		"..", "..", "..", "..",
		"apps", "model-discovery", "prisma", "migrations",
	))
	rowPattern := regexp.MustCompile(`\('([^']+)',\s*'([^']+)',\s*true,`)
	schemaContractSet := make(map[Contract]struct{})
	err := filepath.WalkDir(migrationsRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "migration.sql" {
			return nil
		}

		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, match := range rowPattern.FindAllStringSubmatch(string(raw), -1) {
			schemaContractSet[canonicalContract(match[1], match[2])] = struct{}{}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("read provider schema migrations: %v", err)
	}
	if len(schemaContractSet) == 0 {
		t.Fatal("provider schema migrations contained no runtime-ready rows")
	}

	schemaContracts := make([]Contract, 0, len(schemaContractSet))
	for contract := range schemaContractSet {
		schemaContracts = append(schemaContracts, contract)
	}
	sort.Slice(schemaContracts, func(i, j int) bool {
		if schemaContracts[i].Provider == schemaContracts[j].Provider {
			return schemaContracts[i].AuthMode < schemaContracts[j].AuthMode
		}

		return schemaContracts[i].Provider < schemaContracts[j].Provider
	})

	if registryContracts := Contracts(); !reflect.DeepEqual(registryContracts, schemaContracts) {
		t.Fatalf("validation registry drifted from runtime-ready schemas:\nregistry=%#v\nschemas=%#v", registryContracts, schemaContracts)
	}
}
