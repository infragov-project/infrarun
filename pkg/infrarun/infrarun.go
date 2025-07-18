package infrarun

import "github.com/infragov-project/infrarun/internal/core/tool"

func CreateToolAndReturnName(name string, image string) string {
	return tool.NewTool(name, image).Name
}
