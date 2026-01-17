package test_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
)

type MockPodmanServerSuite struct {
	suite.Suite
	server *test.MockPodmanServer
}

func TestMockPodmanServer(t *testing.T) {
	suite.Run(t, new(MockPodmanServerSuite))
}

func (s *MockPodmanServerSuite) SetupTest() {
	s.server = test.NewMockPodmanServer()
}

func (s *MockPodmanServerSuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *MockPodmanServerSuite) TestStartsAndProvidesURL() {
	s.NotEmpty(s.server.URL())
	s.Contains(s.server.URL(), "http://")
}

func (s *MockPodmanServerSuite) TestRespondsToPingEndpoint() {
	resp, err := http.Get(s.server.URL() + "/_ping")
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *MockPodmanServerSuite) TestRespondsToVersionEndpoint() {
	resp, err := http.Get(s.server.URL() + "/version")
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/json", resp.Header.Get("Content-Type"))
}

func (s *MockPodmanServerSuite) TestRespondsToLibpodVersionEndpoint() {
	resp, err := http.Get(s.server.URL() + "/libpod/version")
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *MockPodmanServerSuite) TestRespondsToVersionedDockerAPI() {
	resp, err := http.Get(s.server.URL() + "/v1.40/version")
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *MockPodmanServerSuite) TestRespondsToInfoEndpoint() {
	resp, err := http.Get(s.server.URL() + "/info")
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *MockPodmanServerSuite) TestReturns404ForUnknownEndpoints() {
	resp, err := http.Get(s.server.URL() + "/unknown/endpoint")
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *MockPodmanServerSuite) TestCapturesRequests() {
	_, err := http.Get(s.server.URL() + "/_ping")
	s.Require().NoError(err)

	_, err = http.Get(s.server.URL() + "/version")
	s.Require().NoError(err)

	requests := s.server.Requests()
	s.GreaterOrEqual(len(requests), 2)

	s.True(s.server.HasRequest("GET", "/_ping"))
	s.True(s.server.HasRequest("GET", "/version"))
}

func (s *MockPodmanServerSuite) TestClearsRequests() {
	_, _ = http.Get(s.server.URL() + "/_ping")
	s.NotEmpty(s.server.Requests())

	s.server.ClearRequests()
	s.Empty(s.server.Requests())
}

func (s *MockPodmanServerSuite) TestCustomHandlerRegistration() {
	s.server.Handle("GET", "/custom/endpoint", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("custom response"))
	})

	resp, err := http.Get(s.server.URL() + "/custom/endpoint")
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusCreated, resp.StatusCode)
}

func (s *MockPodmanServerSuite) TestDualAPIHandlerRegistration() {
	s.server.HandleFunc("GET", "/libpod/containers/json", "/containers/json", func(w http.ResponseWriter, _ *http.Request) {
		test.WriteJSON(w, []test.ContainerListResponse{
			{ID: "abc123", Names: []string{"test-container"}, State: "running"},
		})
	})

	s.Run("libpod path", func() {
		resp, err := http.Get(s.server.URL() + "/libpod/containers/json")
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("docker-compatible path", func() {
		resp, err := http.Get(s.server.URL() + "/containers/json")
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("versioned docker path", func() {
		resp, err := http.Get(s.server.URL() + "/v1.40/containers/json")
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusOK, resp.StatusCode)
	})
}

func (s *MockPodmanServerSuite) TestPathPatternMatchingWithPlaceholders() {
	s.server.Handle("GET", "/libpod/containers/{id}/json", func(w http.ResponseWriter, _ *http.Request) {
		test.WriteJSON(w, test.ContainerInspectResponse{
			ID:   "matched-container",
			Name: "/test",
		})
	})

	resp, err := http.Get(s.server.URL() + "/libpod/containers/abc123/json")
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *MockPodmanServerSuite) TestErrorResponseHelper() {
	s.server.Handle("GET", "/error/endpoint", func(w http.ResponseWriter, _ *http.Request) {
		test.WriteError(w, http.StatusNotFound, "container not found")
	})

	resp, err := http.Get(s.server.URL() + "/error/endpoint")
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *MockPodmanServerSuite) TestResetHandlers() {
	s.server.Handle("GET", "/custom", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	s.Run("custom endpoint works before reset", func() {
		resp, err := http.Get(s.server.URL() + "/custom")
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.server.ResetHandlers()

	s.Run("custom endpoint returns 404 after reset", func() {
		resp, err := http.Get(s.server.URL() + "/custom")
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusNotFound, resp.StatusCode)
	})

	s.Run("base endpoints still work after reset", func() {
		resp, err := http.Get(s.server.URL() + "/_ping")
		s.Require().NoError(err)
		defer func() { _ = resp.Body.Close() }()
		s.Equal(http.StatusOK, resp.StatusCode)
	})
}

type BinaryDetectionSuite struct {
	suite.Suite
}

func TestBinaryDetection(t *testing.T) {
	suite.Run(t, new(BinaryDetectionSuite))
}

func (s *BinaryDetectionSuite) TestIsPodmanAvailable() {
	// This test just verifies the function doesn't panic
	// The result depends on whether podman is installed
	available := test.IsPodmanAvailable()
	s.T().Logf("IsPodmanAvailable: %v", available)
}

func (s *BinaryDetectionSuite) TestIsDockerAvailable() {
	// This test just verifies the function doesn't panic
	// The result depends on whether docker is installed
	available := test.IsDockerAvailable()
	s.T().Logf("IsDockerAvailable: %v", available)
}
