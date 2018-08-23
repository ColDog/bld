package builder

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"
)

func readBuild(filename string) (Build, error) {
	var build Build
	f, err := os.Open(filename)
	if err != nil {
		return build, err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return build, err
	}
	json, err := yaml.YAMLToJSON(data)
	if err != nil {
		return build, err
	}
	err = Validate(json)
	if err != nil {
		return build, err
	}
	err = yaml.Unmarshal(data, &build)
	if err != nil {
		return build, err
	}
	return build, nil
}

func addNamespace(b *Build) {
	const sep = "_"

	for idx, s := range b.Sources {
		s.Name = b.Name + sep + s.Name
		b.Sources[idx] = s
	}
	for idx, s := range b.Volumes {
		s.Name = b.Name + sep + s.Name
		b.Volumes[idx] = s
	}
	for idx, s := range b.Steps {
		s.Name = b.Name + sep + s.Name
		for idx2, i := range s.Imports {
			if !strings.Contains(i.Source, sep) {
				i.Source = b.Name + sep + i.Source
				s.Imports[idx2] = i
			}
		}
		for idx2, i := range s.Exports {
			if !strings.Contains(i.Source, sep) {
				i.Source = b.Name + sep + i.Source
				s.Exports[idx2] = i
			}
		}
		for idx2, i := range s.Volumes {
			if !strings.Contains(i.Source, sep) {
				i.Source = b.Name + sep + i.Source
				s.Volumes[idx2] = i
			}
		}
		b.Steps[idx] = s
	}
}

// Read will read a build.
func Read(filename string) (Build, error) {
	main, err := readBuild(filename)
	if err != nil {
		return main, err
	}

	for _, filename := range main.Requires {
		b, err := readBuild(filename)
		if err != nil {
			return main, err
		}
		bp := &b
		addNamespace(bp)

		main.Volumes = append(main.Volumes, bp.Volumes...)
		main.Steps = append(main.Steps, bp.Steps...)
		main.Sources = append(main.Sources, bp.Sources...)
	}

	return main, err
}
