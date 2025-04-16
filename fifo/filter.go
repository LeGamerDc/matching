package fifo

import "slices"

func (p *PoolProfile) Allow(now int64, t *Ticket) bool {
	for _, f := range p.StringFilters {
		if !f.allow(t) {
			return false
		}
	}
	for _, f := range p.IntFilters {
		if !f.allow(now, t) {
			return false
		}
	}
	for _, f := range p.FloatFilters {
		if !f.allow(t) {
			return false
		}
	}
	return true
}

func findString(args []StringArg, key string) (string, bool) {
	for _, arg := range args {
		if arg.Key == key {
			return arg.Value, true
		}
	}
	return "", false
}

func findInt(args []IntArg, key string) (int64, bool) {
	for _, arg := range args {
		if arg.Key == key {
			return arg.Value, true
		}
	}
	return 0, false
}

func findFloat(args []FloatArg, key string) (float64, bool) {
	for _, arg := range args {
		if arg.Key == key {
			return arg.Value, true
		}
	}
	return 0, false
}

func (f *StringFilter) allow(t *Ticket) bool {
	if f.Op == EqualOp {
		s, ok := findString(t.StringArgs, f.Arg)
		if !ok {
			return false
		}
		return f.Value == s
	}
	if f.Op == NotEqualOp {
		s, ok := findString(t.StringArgs, f.Op)
		if !ok {
			return true
		}
		return f.Value != s
	}
	return true
}

func (f *IntFilter) allow(now int64, t *Ticket) bool {
	var i int64
	switch f.Arg {
	case "$wait":
		i = now - t.startMatch
	case "$member":
		i = int64(len(t.Members))
	default:
		var ok bool
		if i, ok = findInt(t.IntArgs, f.Arg); !ok {
			return false
		}
	}
	return i >= f.Min && i <= f.Max && slices.Index(f.Excludes, i) < 0
}

func (f *FloatFilter) allow(t *Ticket) bool {
	i, ok := findFloat(t.FloatArgs, f.Arg)
	if !ok {
		return false
	}
	return i >= f.Min && i <= f.Max
}
