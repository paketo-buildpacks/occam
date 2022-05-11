package occam

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/paketo-buildpacks/packit/v2/pexec"
)

func NewVenom() Venom {
	return Venom{
		executable: pexec.NewExecutable("venom"),
		vars:       map[string]string{},
	}

}

type Venom struct {
	executable Executable
	vars       map[string]string
	verbose    bool
}

func (v Venom) WithExecutable(executable Executable) Venom {
	v.executable = executable
	return v
}

func (v Venom) WithVerbose() Venom {
	v.verbose = true
	return v
}

func (v Venom) WithPort(port string) Venom {
	return v.WithNamedPort("port", port)
}

func (v Venom) WithNamedPort(name string, port string) Venom {
	v.vars[name] = port
	return v
}

func (v Venom) WithVar(name string, value string) Venom {
	v.vars[name] = value
	return v
}

func (v Venom) Execute(venomPath string) (string, error) {
	args := []string{"run"}

	if v.verbose {
		args = append(args, "-vv")
	}

	// looping this way to ensure we have a consistent order
	keys := make([]string, 0, len(v.vars))
	for k := range v.vars {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		args = append(args, "--var", fmt.Sprintf("%s=%s", key, v.vars[key]))
	}

	args = append(args, venomPath)

	log := bytes.NewBuffer(nil)
	err := v.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: log,
		Stderr: log,
	})

	if err != nil {
		return log.String(), fmt.Errorf("failed to run venom: %w\n\nOutput:\n%s", err, log)
	}
	return log.String(), nil
}
