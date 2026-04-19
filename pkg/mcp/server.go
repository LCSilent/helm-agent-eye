package mcp

import (
	"slices"

	"github.com/LCSilent/helm-agent-eye/pkg/helm"
	"github.com/LCSilent/helm-agent-eye/pkg/k8s"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	server     *server.MCPServer
	k8s        *k8s.Kubernetes
	helmClient *helm.HelmClient
}

func NewServer(name, version string) (*Server, error) {
	s := &Server{
		server: server.NewMCPServer(
			name,
			version,
			server.WithResourceCapabilities(true, true),
			server.WithPromptCapabilities(true),
			server.WithLogging(),
		),
	}
	k8s, err := k8s.NewKubernetes()
	if err != nil {
		return nil, err
	}
	s.k8s = k8s
	s.helmClient = helm.NewHelmClient(k8s, k8s.RestConfig())

	s.server.AddTools(slices.Concat(
		s.initResource(),
		s.initPod(),
		s.initDeployment(),
		s.initService(),
		s.initStatefulSet(),
		s.initNode(),
		s.initIngress(),
		s.initCronJob(),
		s.initNetworkPolicy(),
		s.initWebhook(),
		s.initHelm(),
		s.initDaemonSet(),
		s.initCRD(),
	)...)

	// test prompt
	s.server.AddPrompt(mcp.NewPrompt("get namespace",
		mcp.WithPromptDescription("get namespaces"),
		mcp.WithArgument("name",
			mcp.ArgumentDescription("the namespace to get"),
		),
	), s.getNamespacePrompt)

	return s, nil
}

func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.server)
}

func (s *Server) ServeSSE() *server.SSEServer {
	options := []server.SSEOption{}
	return server.NewSSEServer(s.server, options...)
}
