package mcp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
)

const updateSnapshotsEnvVar = "UPDATE_SNAPSHOTS"

type McpServerSuite struct {
	test.McpSuite
	updateSnapshots bool
}

func TestMcpServer(t *testing.T) {
	suite.Run(t, new(McpServerSuite))
}

func (s *McpServerSuite) SetupTest() {
	s.McpSuite.SetupTest()
	s.updateSnapshots = os.Getenv(updateSnapshotsEnvVar) != ""
}

func (s *McpServerSuite) TestListTools() {
	expectedTools := []string{
		"container_inspect",
		"container_list",
		"container_logs",
		"container_remove",
		"container_run",
		"container_stop",
		"image_build",
		"image_list",
		"image_pull",
		"image_push",
		"image_remove",
		"network_list",
		"volume_list",
	}

	tools, err := s.ListTools()
	s.Require().NoError(err)

	nameSet := make(map[string]bool)
	for _, tool := range tools.Tools {
		nameSet[tool.Name] = true
	}

	for _, name := range expectedTools {
		s.Run("has "+name+" tool", func() {
			s.True(nameSet[name], "tool %s not found", name)
		})
	}
}

func (s *McpServerSuite) TestToolDefinitionsSnapshot() {
	tools, err := s.ListTools()
	s.Require().NoError(err)

	// Sort tools by name for consistent snapshots
	sort.Slice(tools.Tools, func(i, j int) bool {
		return tools.Tools[i].Name < tools.Tools[j].Name
	})

	s.assertJsonSnapshot("tool_definitions.json", tools.Tools)
}

// assertJsonSnapshot compares actual data against a JSON snapshot file.
// Set UPDATE_SNAPSHOTS=1 environment variable to regenerate snapshot files.
func (s *McpServerSuite) assertJsonSnapshot(snapshotFile string, actual any) {
	_, file, _, _ := runtime.Caller(1)
	snapshotPath := filepath.Join(filepath.Dir(file), "testdata", snapshotFile)
	actualJson, err := json.MarshalIndent(actual, "", "  ")
	s.Require().NoErrorf(err, "failed to marshal actual data: %v", err)
	if s.updateSnapshots {
		err := os.WriteFile(snapshotPath, append(actualJson, '\n'), 0644)
		s.Require().NoErrorf(err, "failed to write snapshot file %s: %v", snapshotFile, err)
		s.T().Logf("Updated snapshot: %s", snapshotFile)
		return
	}
	expectedJson := test.ReadFile("testdata", snapshotFile)
	s.JSONEq(
		expectedJson,
		string(actualJson),
		"snapshot %s does not match - to update snapshots re-run the tests with %s=1",
		snapshotFile,
		updateSnapshotsEnvVar,
	)
}

func (s *McpServerSuite) TestCallToolWithMalformedArguments() {
	s.Run("string instead of object arguments returns error", func() {
		// Send arguments as a JSON string instead of an object
		result, err := s.CallToolRaw("container_list", `"this is a string, not an object"`)
		s.Require().NoError(err, "HTTP request should succeed")
		s.NotNil(result.Error, "should return a JSON-RPC error")
		s.Contains(result.Error.Message, "failed to unmarshal arguments", "error should mention unmarshal failure")
	})

	s.Run("array instead of object arguments returns error", func() {
		// Send arguments as a JSON array instead of an object
		result, err := s.CallToolRaw("container_list", `[1, 2, 3]`)
		s.Require().NoError(err, "HTTP request should succeed")
		s.NotNil(result.Error, "should return a JSON-RPC error")
		s.Contains(result.Error.Message, "failed to unmarshal arguments", "error should mention unmarshal failure")
	})

	s.Run("number instead of object arguments returns error", func() {
		// Send arguments as a JSON number instead of an object
		result, err := s.CallToolRaw("container_list", `42`)
		s.Require().NoError(err, "HTTP request should succeed")
		s.NotNil(result.Error, "should return a JSON-RPC error")
		s.Contains(result.Error.Message, "failed to unmarshal arguments", "error should mention unmarshal failure")
	})
}

func (s *McpServerSuite) TestCallToolWithNullArguments() {
	// Set up mock for container_list
	s.WithContainerList([]test.ContainerListResponse{})

	// Send arguments as null - should be treated as empty arguments
	result, err := s.CallToolRaw("container_list", `null`)
	s.Require().NoError(err, "HTTP request should succeed")
	s.Nil(result.Error, "should not return a JSON-RPC error")
	s.NotNil(result.Result, "should return a result")
}
