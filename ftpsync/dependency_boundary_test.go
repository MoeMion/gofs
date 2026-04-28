package ftpsync_test

import (
	"os/exec"
	"sort"
	"strings"
	"testing"
)

func TestPackageDependencyBoundaryRejectsOldRuntime(t *testing.T) {
	// Keep this exact command text discoverable for the Phase 8 acceptance check.
	const dependencyCommand = "go list -deps ./..."

	cmd := exec.Command("go", "list", "-deps", "./...")
	cmd.Dir = ".."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", dependencyCommand, err, out)
	}

	forbidden := []string{
		"github.com/no-src/gofs/action",
		"github.com/no-src/gofs/api",
		"github.com/no-src/gofs/auth",
		"github.com/no-src/gofs/cmd",
		"github.com/no-src/gofs/conf",
		"github.com/no-src/gofs/core",
		"github.com/no-src/gofs/daemon",
		"github.com/no-src/gofs/driver/minio",
		"github.com/no-src/gofs/driver/sftp",
		"github.com/no-src/gofs/eventlog",
		"github.com/no-src/gofs/flag",
		"github.com/no-src/gofs/logger",
		"github.com/no-src/gofs/monitor",
		"github.com/no-src/gofs/report",
		"github.com/no-src/gofs/server",
		"github.com/no-src/gofs/sync",
		"github.com/no-src/gofs/task",
		"github.com/gin-contrib",
		"github.com/gin-gonic/gin",
		"github.com/minio/minio-go/v7",
		"github.com/no-src/nscache",
		"github.com/pkg/sftp",
		"github.com/quic-go/quic-go",
		"golang.org/x/oauth2",
		"google.golang.org/grpc",
		"google.golang.org/protobuf",
	}

	deps := strings.Split(strings.TrimSpace(string(out)), "\n")
	matches := make(map[string][]string)
	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}
		for _, fragment := range forbidden {
			if dep == fragment || strings.HasPrefix(dep, fragment+"/") || strings.Contains(dep, fragment) {
				matches[fragment] = append(matches[fragment], dep)
			}
		}
	}

	if len(matches) == 0 {
		return
	}

	fragments := make([]string, 0, len(matches))
	for fragment := range matches {
		fragments = append(fragments, fragment)
	}
	sort.Strings(fragments)

	var report strings.Builder
	report.WriteString("ftpsync dependency graph includes forbidden old runtime dependencies:\n")
	report.WriteString("\nForbidden fragment | Matched dependency\n")
	report.WriteString("--- | ---\n")
	for _, fragment := range fragments {
		sort.Strings(matches[fragment])
		for _, dep := range matches[fragment] {
			report.WriteString(fragment)
			report.WriteString(" | ")
			report.WriteString(dep)
			report.WriteByte('\n')
		}
	}

	t.Fatal(report.String())
}
