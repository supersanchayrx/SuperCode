package tools

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Execute     func(args map[string]interface{}) (string, error)
}

type Registry struct {
	tools map[string]Tool
}

// this returns a pointer to the a registry (typical constructor)
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool), //creating this registry and returning its address
	}
}

// registering the tool to the registry r
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name] = tool
}

func (r *Registry) GetToolByName(toolName string) (Tool, bool) {
	tool, exists := r.tools[toolName]
	return tool, exists
}

func (r *Registry) GetAllTools() []Tool {
	all := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		all = append(all, tool)
	}

	return all
}
