package graph

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/coldog/bld/pkg/builder"
	"github.com/stretchr/testify/require"
)

func test(t *testing.T, s *Solver) []string {
	s.Solve()
	defer s.Close()

	c := make(chan string)

	wg := &sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wg.Done()

			for {
				id, err := s.Select(context.Background())
				if err != nil {
					break
				}
				println(id)
				c <- id
				if i%2 == 0 {
					time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
				}
				s.Done(id)
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	out := []string{}
	for k := range c {
		out = append(out, k)
	}
	return out
}

func TestSolver_Graph(t *testing.T) {
	out := test(t, &Solver{
		Build: builder.Build{
			Name: "test",
			Sources: []builder.Source{
				{Name: "r1", Target: "/tmp/"},
			},
			Steps: []builder.Step{
				{Name: "s1", Imports: []builder.Mount{
					{Source: "r1", Mount: "/usr/src/app"},
				}, Exports: []builder.Mount{
					{Source: "r2", Mount: "/usr/src/app2"},
				}},
				{Name: "s2", Imports: []builder.Mount{
					{Source: "r2", Mount: "/usr/src/app"},
				}},
			},
		},
	})
	require.ElementsMatch(t, []string{"source/r1", "s1", "s2"}, out)
}

func TestSolver_Simple(t *testing.T) {
	out := test(t, &Solver{
		Build: builder.Build{
			Name: "test",
			ID:   "test-1",
			Sources: []builder.Source{
				{Name: "r1", Target: "/tmp"},
			},
			Steps: []builder.Step{
				{
					Name:    "s1-1",
					Imports: []builder.Mount{{Source: "r1", Mount: "/usr/src/app"}},
				},
				{
					Name:    "s1-2",
					Imports: []builder.Mount{{Source: "r1", Mount: "/usr/src/app"}},
				},
				{
					Name:    "s1-3",
					Imports: []builder.Mount{{Source: "r1", Mount: "/usr/src/app"}},
				},
			},
		},
	})
	require.ElementsMatch(t, []string{"source/r1", "s1-1", "s1-2", "s1-3"}, out)
}
