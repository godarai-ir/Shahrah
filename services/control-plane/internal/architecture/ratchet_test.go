// Package architecture protects control-plane dependency and maintainability boundaries.
package architecture

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	maxStoreReceiverMethods                = 1
	maxAdminPersistenceImports             = 0
	maxProxyPersistenceImports             = 0
	maxDirectEnvironmentReadsOutsideConfig = 0
	maxOversizedProductionFiles            = 28
	maxFocusedRepositoryMethods            = 25
	maxApplicationPortMethods              = 10
	oversizedFileLineThreshold             = 800
)

func TestStoreCompatibilityFacadeCannotGrow(t *testing.T) {
	t.Parallel()

	count := countReceiverMethods(t, filepath.Join(moduleRoot(t), "internal", "storage"), "Store")
	require.LessOrEqual(t, count, maxStoreReceiverMethods, "move feature persistence to a repository and lower the ratchet")
}

func TestAdminPersistenceImportsCannotGrow(t *testing.T) {
	t.Parallel()

	count := countImports(t, filepath.Join(moduleRoot(t), "internal", "admin"), map[string]struct{}{
		"github.com/Delpak/shahrah/services/control-plane/internal/storage": {},
		"github.com/jackc/pgx/v5":        {},
		"github.com/jackc/pgx/v5/pgconn": {},
	})
	require.LessOrEqual(t, count, maxAdminPersistenceImports, "translate persistence errors behind application ports and lower the ratchet")
}

func TestProxyPersistenceImportsCannotGrow(t *testing.T) {
	t.Parallel()

	count := countImports(t, filepath.Join(moduleRoot(t), "internal", "proxy"), map[string]struct{}{
		"github.com/Delpak/shahrah/services/control-plane/internal/storage": {},
		"github.com/jackc/pgx/v5":        {},
		"github.com/jackc/pgx/v5/pgconn": {},
	})
	require.LessOrEqual(t, count, maxProxyPersistenceImports, "proxy adapters must depend on application ports")
}

func TestLegacyGodRepositoriesHaveNoReceiverMethods(t *testing.T) {
	t.Parallel()

	root := filepath.Join(moduleRoot(t), "internal", "storage")
	for _, receiver := range []string{"PlanRepository", "WalletRepository", "ObservabilityRepository"} {
		count := countReceiverMethods(t, root, receiver)
		require.Zero(t, count, "%s must remain a capability composition", receiver)
	}
}

func TestFocusedRepositoriesStayFocused(t *testing.T) {
	t.Parallel()

	root := filepath.Join(moduleRoot(t), "internal", "storage")

	repositories := []string{
		"PlanCatalogRepository",
		"SubscriptionRepository",
		"EntitlementRepository",
		"WalletLedgerRepository",
		"ExchangeRateRepository",
		"WalletPaymentRepository",
		"FinancialRepository",
		"ObservabilityHealthRepository",
		"TraceRepository",
		"RequestLogRepository",
		"LogManagementRepository",
		"MetricsRepository",
		"SpendRepository",
		"ObservabilitySettingsRepository",
		"TokenUsageRepository",
	}
	for _, receiver := range repositories {
		count := countReceiverMethods(t, root, receiver)
		require.LessOrEqual(t, count, maxFocusedRepositoryMethods, "%s must be split by capability", receiver)
	}
}

func TestApplicationServicesDoNotEmbedRepositories(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	for _, packageName := range []string{"iam", "policies", "provideraccounts", "tenantadmin", "virtualmodels"} {
		file := filepath.Join(root, "internal", packageName, "service.go")
		parsed := parseGoFile(t, file)
		ast.Inspect(parsed, func(node ast.Node) bool {
			typeSpec, ok := node.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "Service" {
				return true
			}

			service, ok := typeSpec.Type.(*ast.StructType)
			require.True(t, ok, "%s.Service must be a struct", packageName)

			for _, field := range service.Fields.List {
				require.NotEmpty(t, field.Names, "%s.Service must use named private ports instead of embedding repositories", packageName)
			}

			return false
		})
	}
}

func TestApplicationPortsStayUseCaseFocused(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	for _, packageName := range []string{"iam", "policies", "provideraccounts", "tenantadmin", "virtualmodels"} {
		file := filepath.Join(root, "internal", packageName, "service.go")
		parsed := parseGoFile(t, file)

		for _, declaration := range parsed.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}

			for _, specification := range general.Specs {
				typeSpec, ok := specification.(*ast.TypeSpec)
				if !ok {
					continue
				}

				port, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}

				methodCount := 0
				for _, method := range port.Methods.List {
					methodCount += len(method.Names)
				}

				require.LessOrEqual(t, methodCount, maxApplicationPortMethods, "%s.%s must be split by use case", packageName, typeSpec.Name.Name)
			}
		}
	}
}

func TestEnvironmentReadsStayAtConfigurationEdge(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	count := 0

	for _, directory := range []string{filepath.Join(root, "internal"), filepath.Join(root, "cmd")} {
		for _, file := range productionGoFiles(t, directory) {
			relative, err := filepath.Rel(root, file)
			require.NoError(t, err)

			if strings.HasPrefix(relative, filepath.Join("internal", "config")) ||
				strings.HasPrefix(relative, filepath.Join("internal", "bootstrap")) {
				continue
			}

			count += countOSEnvironmentCalls(t, file)
		}
	}

	require.LessOrEqual(t, count, maxDirectEnvironmentReadsOutsideConfig, "load settings in internal/config, inject them, and lower the ratchet")

	for _, packageName := range []string{"admin", "storage", "router", "runtimeprofile"} {
		packageCount := 0
		for _, file := range productionGoFiles(t, filepath.Join(root, "internal", packageName)) {
			packageCount += countOSEnvironmentCalls(t, file)
		}

		require.Zero(t, packageCount, "%s must consume immutable injected configuration", packageName)
	}
}

func TestRuntimePackagesCannotReadEnvironmentEdge(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	count := 0

	for _, file := range productionGoFiles(t, filepath.Join(root, "internal")) {
		relative, err := filepath.Rel(root, file)
		require.NoError(t, err)

		if strings.HasPrefix(relative, filepath.Join("internal", "config")) ||
			strings.HasPrefix(relative, filepath.Join("internal", "bootstrap")) {
			continue
		}

		parsed := parseGoFile(t, file)
		for _, imported := range parsed.Imports {
			path, unquoteErr := strconv.Unquote(imported.Path.Value)
			require.NoError(t, unquoteErr)

			if path == "github.com/Delpak/shahrah/services/control-plane/internal/config/environment" {
				count++
			}
		}
	}

	require.Zero(t, count, "runtime packages must receive typed configuration through bootstrap")
}

func TestOversizedProductionFilesCannotGrow(t *testing.T) {
	t.Parallel()

	count := 0

	for _, file := range productionGoFiles(t, filepath.Join(moduleRoot(t), "internal")) {
		if lineCount(t, file) > oversizedFileLineThreshold {
			count++
		}
	}

	require.LessOrEqual(t, count, maxOversizedProductionFiles, "split files by use case and lower the ratchet")
}

func moduleRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)

	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

func productionGoFiles(t *testing.T, root string) []string {
	t.Helper()

	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}

		files = append(files, path)

		return nil
	})

	require.NoError(t, err)

	return files
}

func countReceiverMethods(t *testing.T, root string, receiverName string) int {
	t.Helper()

	count := 0

	for _, file := range productionGoFiles(t, root) {
		parsed := parseGoFile(t, file)
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv == nil || len(function.Recv.List) == 0 {
				continue
			}

			if receiverTypeName(function.Recv.List[0].Type) == receiverName {
				count++
			}
		}
	}

	return count
}

func receiverTypeName(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return receiverTypeName(value.X)
	default:
		return ""
	}
}

func countImports(t *testing.T, root string, prohibited map[string]struct{}) int {
	t.Helper()

	count := 0

	for _, file := range productionGoFiles(t, root) {
		parsed := parseGoFile(t, file)
		for _, imported := range parsed.Imports {
			path, err := strconv.Unquote(imported.Path.Value)
			require.NoError(t, err)

			if _, found := prohibited[path]; found {
				count++
			}
		}
	}

	return count
}

func countOSEnvironmentCalls(t *testing.T, file string) int {
	t.Helper()

	count := 0

	ast.Inspect(parseGoFile(t, file), func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || (selector.Sel.Name != "Getenv" && selector.Sel.Name != "LookupEnv") {
			return true
		}

		packageName, ok := selector.X.(*ast.Ident)
		if ok && packageName.Name == "os" {
			count++
		}

		return true
	})

	return count
}

func parseGoFile(t *testing.T, file string) *ast.File {
	t.Helper()

	parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.SkipObjectResolution)
	require.NoError(t, err)

	return parsed
}

func lineCount(t *testing.T, file string) int {
	t.Helper()

	handle, err := os.Open(file)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, handle.Close())
	}()

	scanner := bufio.NewScanner(handle)
	count := 0

	for scanner.Scan() {
		count++
	}

	require.NoError(t, scanner.Err())

	return count
}
