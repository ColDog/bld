package template

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	gotemplate "text/template"
)

func sh(arg string, args ...string) *string {
	out, err := exec.Command(arg, args...).CombinedOutput()
	if err != nil {
		return nil
	}
	s := strings.TrimSpace(string(out))
	return &s
}

func gitMap() map[string]*string {
	return map[string]*string{
		"ShaShort": sh("git", "rev-parse", "--short", "HEAD"),
		"Sha":      sh("git", "rev-parse", "HEAD"),
		"Branch":   sh("git", "rev-parse", "--abbrev-ref", "HEAD"),
	}
}

func environMap() map[string]string {
	m := map[string]string{}
	for _, val := range os.Environ() {
		spl := strings.Split(val, "=")
		key := spl[0]
		val := strings.Join(spl[1:], "=")
		m[key] = val
	}
	return m
}

// Struct template.
func Struct(i interface{}) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	tpl, err := gotemplate.New("").Parse(string(data))
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	input := struct {
		Git     map[string]*string
		Environ map[string]string
	}{
		Git:     gitMap(),
		Environ: environMap(),
	}
	err = tpl.Execute(buf, input)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf.Bytes(), i)
}
