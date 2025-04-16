package mwm

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	b := New(10)
	for i := 1; i <= 10; i++ {
		for j := i + 1; j <= 10; j++ {
			b.AddEdge(i, j, i)
		}
	}
	b.initialize()
	for i := 1; i <= 10; i++ {
		x, y := b.ofs[i], b.ofs[i+1]
		for j := x; j < y; j++ {
			assert.Equal(t, i, b.edges[j].from)
		}
	}
}

func TestSolve1(t *testing.T) {
	b := New(4)
	b.AddEdge(1, 2, 1)
	b.AddEdge(3, 4, 1)
	b.AddEdge(1, 3, 1)
	b.AddEdge(2, 4, 2)
	pair, rest, w := b.Solve()
	fmt.Println(pair, rest, w)
}

func TestSolve2(t *testing.T) {
	b := New(4)
	b.AddEdge(1, 2, 1)
	b.AddEdge(2, 3, 10)
	b.AddEdge(3, 4, 1)
	fmt.Println(b.Solve())
}

const p = 0.17

func BenchmarkB5(b *testing.B) {
	n := b.N
	b5 := New(n)
	pp := p
	if n < 100 {
		pp = 1
	}
	for i := 1; i <= n; i++ {
		for j := i + 1; j <= n; j++ {
			if rand.Float64() < pp {
				w := rand.N(300) + 1
				b5.AddEdge(i, j, w)
			}
		}
	}
	b.ResetTimer()
	s := time.Now()
	b5.Solve()
	fmt.Printf("bench %d: %v\n", n, time.Since(s))
}
