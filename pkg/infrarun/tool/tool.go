package tool

import "github.com/infragov-project/infrarun/internal/core/tools"

// A Tool represents the concept of a code analysis tool running inside a [Docker] container,
// together with a parser that converts its output into [SARIF].
//
// [Docker]: https://www.docker.com/
// [SARIF]: https://sarifweb.azurewebsites.net/
type Tool struct {
	Impl *tools.Tool
}

// Name returns the display name of the given tool.
func (t Tool) Name() string {
	return t.Impl.Name
}

// Image returns the [Docker image reference] of the image used by the given tool.
//
// [Docker image reference]: https://docs.docker.com/reference/cli/docker/image/tag/#description
func (t Tool) Image() string {
	return t.Impl.Image
}

func ToolFromImpl(impl *tools.Tool) Tool {
	return Tool{impl}
}

type ToolInstance struct {
	Impl *tools.ToolInstance
}

func ToolInstanceFromImpl(impl *tools.ToolInstance) ToolInstance {
	return ToolInstance{impl}
}

// GetAvailableTools returns a map with all the infrarun [Tool] available in the current process.
// This map has the [Tool]'s display name as the keys.
func GetAvailableTools() map[string]Tool {
	impls := tools.GetEmbedToolDefinitions()

	t := make(map[string]Tool)

	for k, v := range impls {
		t[k] = ToolFromImpl(&v)
	}

	return t
}
