package tool

type Tool struct {
	Name  string
	Image string
}

func NewTool(name string, image string) *Tool {
	return &Tool{
		name,
		image,
	}
}
