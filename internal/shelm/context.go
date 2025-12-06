package shelm

// Context is the execution context passed to templates as the dot (.) value.
type Context struct {
	Vars map[string]any `json:"vars" yaml:"vars"`
}

// NewContext creates a new Context with an initialized Vars map.
func NewContext() *Context {
	return &Context{
		Vars: make(map[string]any),
	}
}

// Set sets a variable in the context.
func (c *Context) Set(name string, value any) {
	c.Vars[name] = value
}

// Get retrieves a variable from the context.
func (c *Context) Get(name string) (any, bool) {
	v, ok := c.Vars[name]
	return v, ok
}

// Delete removes a variable from the context.
func (c *Context) Delete(name string) {
	delete(c.Vars, name)
}

// Merge merges a map into the context's Vars (shallow merge).
func (c *Context) Merge(data map[string]any) {
	for k, v := range data {
		c.Vars[k] = v
	}
}
