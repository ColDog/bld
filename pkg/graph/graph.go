package graph

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/coldog/bld/pkg/builder"
)

// ErrFinished will be returned if the Solver is closed.
var ErrFinished = errors.New("selector finished")

// Solver is a graph solver. It takes a given build and returns units of work to
// goroutines.
type Solver struct {
	Build   builder.Build
	Workers int

	dependencies map[string]set
	complete     *watchSet
	selector     chan string
	done         chan struct{}
}

// Close should be called when processing is finished.
func (s *Solver) Close() {
	if s.complete != nil {
		s.complete.close()
	}
	s.dependencies = nil
	if s.selector != nil {
		s.selector = nil
	}
	if s.done != nil {
		close(s.done)
	}
}

// Done marks a unit of work as complete, any dependencies will now be
// available.
func (s *Solver) Done(id string) { s.complete.add(id) }

// Select will select work that needs to be completed from the graph.
func (s *Solver) Select(ctx context.Context) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-s.done:
			cancel()
			return
		case <-ctx.Done():
			return
		}
	}()

	var id string
	select {
	case selID, ok := <-s.selector:
		if !ok {
			return "", ErrFinished
		}
		id = selID
	case <-ctx.Done():
		return "", ctx.Err()
	}
	for dep := range s.dependencies[id] {
		if err := s.complete.wait(ctx, dep); err != nil {
			return "", err
		}
	}
	return id, nil
}

// Solve will begin the solving process. Select can be called after this to
// select work that needs to be completed.
func (s *Solver) Solve() {
	s.selector = make(chan string)
	s.complete = newWatchSet()
	s.done = make(chan struct{})

	// Build an adjacency list.
	sourceToStep := map[string]string{}
	adjacency := map[string]set{}
	for _, s := range s.Build.Steps {
		adjacency[s.Name] = set{}
		for _, e := range s.Exports {
			sourceToStep[e.Source] = s.Name
		}
	}
	for _, s := range s.Build.Sources {
		name := "source/" + s.Name
		sourceToStep[s.Name] = name
		adjacency[name] = set{}
	}

	// Mapping from step name to the next step.
	for _, s := range s.Build.Steps {
		for _, src := range s.Imports {
			adj := sourceToStep[src.Source]
			adjacency[adj].add(s.Name)
		}
	}

	// Mapping from step name to those that point to it.
	dependencies := map[string]set{}
	for key := range adjacency {
		dependencies[key] = set{}
		for parent, adjacent := range adjacency {
			if adjacent.has(key) {
				dependencies[key].add(parent)
			}
		}
	}

	// Find start steps as all nodes with no incoming edges.
	stack := []string{}
	for key, parents := range dependencies {
		if len(parents) > 0 {
			continue
		}
		stack = append(stack, key)
	}

	discovered := map[string]bool{}
	s.dependencies = dependencies

	go func() {
		defer close(s.selector)

		for len(stack) > 0 {
			select {
			case <-s.done:
				return
			default:
			}

			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if !discovered[v] {
				discovered[v] = true
				s.selector <- v
				for _, edge := range adjacency[v].list() {
					stack = append(stack, edge)
				}
			}
		}
	}()
}

type set map[string]bool

func (s set) add(item string)      { s[item] = true }
func (s set) has(item string) bool { return s[item] }

func (s set) list() (l []string) {
	for key := range s {
		l = append(l, key)
	}
	sort.Strings(l)
	return l
}

func newWatchSet() *watchSet {
	return &watchSet{
		m:    map[string]bool{},
		subs: map[string]map[chan string]struct{}{},
	}
}

type watchSet struct {
	sync.Mutex
	m    map[string]bool
	subs map[string]map[chan string]struct{}
}

func (s *watchSet) add(item string) {
	s.Lock()
	s.m[item] = true
	if s.subs[item] != nil {
		for c := range s.subs[item] {
			c <- item
		}
	}
	s.Unlock()
}

func (s *watchSet) wait(ctx context.Context, item string) error {
	s.Lock()
	status := s.m[item]
	s.Unlock()
	if status {
		return nil
	}

	c := make(chan string)

	s.Lock()
	if s.subs[item] == nil {
		s.subs[item] = map[chan string]struct{}{}
	}
	s.subs[item][c] = struct{}{}
	s.Unlock()

	select {
	case <-c:
		s.Lock()
		delete(s.subs[item], c)
		s.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *watchSet) close() {
	s.m = nil
	for _, subs := range s.subs {
		for sub := range subs {
			close(sub)
		}
	}
}
