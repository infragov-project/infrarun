package plan

import (
	"github.com/infragov-project/infrarun/internal/core/engine"
	"github.com/infragov-project/infrarun/pkg/infrarun/tool"
)

type Plan struct {
	Runs []*Run
}

type Run struct {
	Impl *engine.ToolExecution
}

func NewRun(path, glob string, tool *tool.Tool, options map[string]any) (*Run, error) {
	instance, err := tool.Impl.ToInstance(options)

	if err != nil {
		return nil, err
	}

	exec, err := engine.NewToolExecution(instance, path, glob)

	if err != nil {
		return nil, err
	}

	return &Run{
		Impl: exec,
	}, nil
}

func NewRunWithDefaultOptions(path, glob string, tool *tool.Tool) (*Run, error) {
	impl, err := tool.Impl.DefaultInstance()

	if err != nil {
		return nil, err
	}

	exec, err := engine.NewToolExecution(impl, path, glob)

	if err != nil {
		return nil, err
	}

	return &Run{
		Impl: exec,
	}, nil
}

func NewSimpleRun(path string, tool *tool.Tool) (*Run, error) {
	return NewRunWithDefaultOptions(path, "**/*", tool)
}

func (p *Plan) AddRun(run *Run) {
	p.Runs = append(p.Runs, run)
}

func (r *Run) ToolName() string {
	return r.Impl.Tool.Name
}
