package analysis

import "sort"

type interval struct{ s, e int64 }

func (i interval) length() int64 {
	if i.e > i.s {
		return i.e - i.s
	}
	return 0
}

// mergeIntervals ordena y fusiona intervalos solapados o contiguos.
func mergeIntervals(iv []interval) []interval {
	if len(iv) == 0 {
		return nil
	}
	c := append([]interval(nil), iv...)
	sort.Slice(c, func(i, j int) bool { return c[i].s < c[j].s })
	out := []interval{c[0]}
	for _, x := range c[1:] {
		last := &out[len(out)-1]
		if x.s <= last.e {
			if x.e > last.e {
				last.e = x.e
			}
			continue
		}
		out = append(out, x)
	}
	return out
}

func totalLen(iv []interval) int64 {
	var t int64
	for _, x := range iv {
		t += x.length()
	}
	return t
}

// intersect devuelve a ∩ b (ambos fusionados y ordenados).
func intersect(a, b []interval) []interval {
	var out []interval
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		s, e := max64(a[i].s, b[j].s), min64(a[i].e, b[j].e)
		if e > s {
			out = append(out, interval{s, e})
		}
		if a[i].e < b[j].e {
			i++
		} else {
			j++
		}
	}
	return out
}

// subtract devuelve a \ b (ambos fusionados y ordenados).
func subtract(a, b []interval) []interval {
	var out []interval
	for _, x := range a {
		cur := x.s
		for _, y := range b {
			if y.e <= cur || y.s >= x.e {
				continue
			}
			if y.s > cur {
				out = append(out, interval{cur, y.s})
			}
			if y.e > cur {
				cur = y.e
			}
		}
		if cur < x.e {
			out = append(out, interval{cur, x.e})
		}
	}
	return out
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
