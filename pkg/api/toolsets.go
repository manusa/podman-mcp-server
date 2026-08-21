package api

type Toolset interface {
	GetName() string
	GetDescription() string
	GetTools() []ServerTool
}
