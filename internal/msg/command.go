//go:generate easyjson -all

package msg

type Command struct {
	Name string   `json:"name,omitempty"`
	Args []string `json:"args,omitempty"`
}

func C(name string, args ...string) *Command {
	return &Command{Name: name, Args: args}
}
