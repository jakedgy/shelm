package shelm

// Context is the execution context passed to templates as the dot (.) value.
type Context struct {
	Values map[string]any `json:"values" yaml:"values"`
}

// NewContext creates a new Context with an initialized Values map.
func NewContext() *Context {
	return &Context{
		Values: make(map[string]any),
	}
}

// Set sets a variable in the context.
func (c *Context) Set(name string, value any) {
	c.Values[name] = value
}

// Get retrieves a variable from the context.
func (c *Context) Get(name string) (any, bool) {
	v, ok := c.Values[name]
	return v, ok
}

// Delete removes a variable from the context.
func (c *Context) Delete(name string) {
	delete(c.Values, name)
}

// Merge merges a map into the context's Values (shallow merge).
func (c *Context) Merge(data map[string]any) {
	for k, v := range data {
		c.Values[k] = v
	}
}
