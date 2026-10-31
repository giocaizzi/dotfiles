// Package udiff renders unified diffs of two texts.
package udiff

import (
	"fmt"
	"strings"
)

const context = 3

// Diff returns the unified diff of a and b, or "" when they are equal.
func Diff(nameA, nameB, a, b string) string {
	x, y := lines(a), lines(b)
	ops := edits(x, y)
	var out strings.Builder
	changed := false
	for i := 0; i < len(ops); {
		if ops[i].kind == ' ' {
			i++
			continue
		}
		// A hunk: changes with up to 2*context unchanged lines between them.
		start := max(i-context, 0)
		end, lastChange := i, i
		for end < len(ops) && end-lastChange <= 2*context {
			if ops[end].kind != ' ' {
				lastChange = end
			}
			end++
		}
		end = min(lastChange+context+1, len(ops))
		if !changed {
			fmt.Fprintf(&out, "--- %s\n+++ %s\n", nameA, nameB)
			changed = true
		}
		hunk(&out, ops[start:end])
		i = end
	}
	return out.String()
}

type op struct {
	kind       byte // ' ', '-' or '+'
	text       string
	aLine, bLn int // 1-based line numbers in a and b
}

func hunk(out *strings.Builder, ops []op) {
	var na, nb int
	for _, o := range ops {
		if o.kind != '+' {
			na++
		}
		if o.kind != '-' {
			nb++
		}
	}
	a0, b0 := ops[0].aLine, ops[0].bLn
	fmt.Fprintf(out, "@@ -%d,%d +%d,%d @@\n", start(a0, na), na, start(b0, nb), nb)
	for _, o := range ops {
		out.WriteByte(o.kind)
		out.WriteString(o.text)
		out.WriteByte('\n')
	}
}

func start(line, n int) int {
	if n == 0 {
		return line - 1
	}
	return line
}

func lines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// edits is the line-level edit script from x to y, by longest common subsequence.
func edits(x, y []string) []op {
	n, m := len(x), len(y)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if x[i] == y[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var ops []op
	i, j := 0, 0
	for i < n || j < m {
		switch {
		case i < n && j < m && x[i] == y[j]:
			ops = append(ops, op{' ', x[i], i + 1, j + 1})
			i++
			j++
		case j < m && (i == n || lcs[i][j+1] >= lcs[i+1][j]):
			ops = append(ops, op{'+', y[j], i + 1, j + 1})
			j++
		default:
			ops = append(ops, op{'-', x[i], i + 1, j + 1})
			i++
		}
	}
	return ops
}
