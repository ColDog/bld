package runner

import "sync"

func newSources() *sources {
	return &sources{
		sources: map[string]*source{},
	}
}

type source struct {
	localDir string
	digest   string
}

type sources struct {
	sync.Mutex

	sources map[string]*source
}

func (sr *sources) add(name string) {
	sr.Lock()
	defer sr.Unlock()
	sr.sources[name] = &source{}
}

func (sr *sources) setDir(name, dir string) {
	sr.Lock()
	defer sr.Unlock()
	if _, ok := sr.sources[name]; !ok {
		sr.sources[name] = &source{}
	}
	sr.sources[name].localDir = dir
}

func (sr *sources) setDigest(name, dir string) {
	sr.Lock()
	defer sr.Unlock()
	if _, ok := sr.sources[name]; !ok {
		sr.sources[name] = &source{}
	}
	sr.sources[name].digest = dir
}

func (sr *sources) getDir(name string) string {
	sr.Lock()
	defer sr.Unlock()
	if _, ok := sr.sources[name]; !ok {
		return ""
	}
	return sr.sources[name].localDir
}

func (sr *sources) getDigest(name string) string {
	sr.Lock()
	defer sr.Unlock()
	if _, ok := sr.sources[name]; !ok {
		return ""
	}
	return sr.sources[name].digest
}

func (sr *sources) dirMap() map[string]string {
	sr.Lock()
	defer sr.Unlock()
	m := map[string]string{}
	for k, v := range sr.sources {
		m[k] = v.localDir
	}
	return m
}
