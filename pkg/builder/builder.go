package builder

type Build struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Volumes []Source `json:"volumes"`
	Sources []Source `json:"sources"`
	Steps   []Step   `json:"steps"`
}

func (b Build) Source(name string) (Source, bool) {
	for _, src := range b.Sources {
		if src.Name == name {
			return src, true
		}
	}
	return Source{}, false
}

func (b Build) Step(name string) (Step, bool) {
	for _, step := range b.Steps {
		if step.Name == name {
			return step, true
		}
	}
	return Step{}, false
}

type Source struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

type Mount struct {
	Source string `json:"source"`
	Mount  string `json:"mount"`
}

type Step struct {
	Name     string   `json:"name"`
	Imports  []Mount  `json:"imports"`
	Volumes  []Mount  `json:"volumes"`
	Exports  []Mount  `json:"exports"`
	Image    string   `json:"image"`
	Commands []string `json:"commands"`
	Tag      string   `json:"tag"`
	Workdir  string   `json:"workdir"`
	Env      []string `json:"env"`
}

type StepExec struct {
	Step

	BuildID    string
	BuildDir   string
	RootDir    string
	SourceDirs map[string]string
}
