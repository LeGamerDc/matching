package fifo

import (
	"slices"
	"sort"
)

type candidate struct {
	tickets   []*Ticket
	me        []int64
	hates     []int64
	anti      []string
	low, high int
	used      bool
}

func (c *candidate) matchTeam(cs []*candidate) bool {
	for _, c2 := range cs {
		for _, id := range c.anti {
			if slices.Index(c2.anti, id) >= 0 {
				return false
			}
		}
	}
	return true
}

func (c *candidate) allow(t *Ticket) bool {
	for _, m := range t.Members {
		if slices.Index(c.hates, m.BlackId) >= 0 {
			return false
		}
	}
	if len(t.BlackList) > 0 {
		for _, id := range c.me {
			if slices.Index(t.BlackList, id) >= 0 {
				return false
			}
		}
	}
	return true
}

func (c *candidate) join(t *Ticket) {
	t.used = true
	for _, m := range t.Members {
		if m.Sort == 0 {
			c.low++
			c.high++
		} else {
			c.high++
		}
		c.me = appendUnique(c.me, m.BlackId)
	}
	for _, id := range t.BlackList {
		c.hates = appendUnique(c.hates, id)
	}
	c.tickets = append(c.tickets, t)
}

func (c *candidate) result(name string, n int) TeamResult {
	tr := TeamResult{
		TeamName: name,
		TicketId: make([]string, 0, len(c.tickets)),
		Members:  make([]MemberResult, 0, n),
	}
	tt := 0
	for _, t := range c.tickets {
		tr.TicketId = append(tr.TicketId, t.TicketId)
		tt += len(t.Members)
		for _, m := range t.Members {
			tr.Members = append(tr.Members, MemberResult{
				MemberId: m.MemberId,
				Extra:    m.Extra,
				sort:     m.Sort,
			})
		}
	}
	if tt <= n {
		return tr
	}
	sort.Slice(tr.Members, func(i, j int) bool {
		return tr.Members[i].sort < tr.Members[j].sort
	})
	tr.CutMembers = tr.Members[n:]
	tr.Members = tr.Members[:n]
	return tr
}

func matchResult(pool PoolProfile, teams []*candidate) MatchResult {
	n := min(len(pool.Teams), len(teams))
	mr := MatchResult{
		PoolName: pool.Name,
		Teams:    make([]TeamResult, 0, n),
	}
	for i := 0; i < n; i++ {
		mr.Teams = append(mr.Teams, teams[i].result(pool.Teams[i], pool.TeamMembers))
	}
	return mr
}

func FifoMatch(pool PoolProfile, tickets map[string]*Ticket, now int64, r ResultSubmitter) {
	n := pool.TeamMembers
	m := len(pool.Teams)
	queue := sortTicket(tickets, n)
	if quickFail(queue, n, m) {
		return
	}
	var cans []*candidate
	// step 1. match inside team
MATCH:
	for i := n; i >= 1; i-- {
		for j := range queue[i] {
			if queue[i][j].used {
				continue
			}
			buf := new(candidate)
			buf.join(queue[i][j])
			ok := search(buf, queue, n, pool.AllowCut, pool.AllowHate)
			if ok {
				cans = append(cans, buf)
				if len(cans) >= pool.MaxMatchPerRound {
					break MATCH
				}
			}
		}
		removeUsed(queue)
	}
	if len(cans) < m {
		return
	}
	// step 2. match between teams
	if m == 1 {
		for _, can := range cans {
			r(MatchResult{
				PoolName: pool.Name,
				Teams:    []TeamResult{can.result(pool.Teams[0], pool.TeamMembers)},
			})
		}
		return
	}

	if pool.BetweenTeamAntiAffinity != "" {
		for _, can := range cans {
			for _, t := range can.tickets {
				if tag, ok := findString(t.StringArgs, pool.BetweenTeamAntiAffinity); ok {
					can.anti = appendUnique(can.anti, tag)
				}
			}
		}
	}
	for i := range cans {
		if rr, ok := searchTeam(cans, i, m); ok {
			r(matchResult(pool, rr))
		}
	}
}

func searchTeam(queue []*candidate, i, n int) (buf []*candidate, ok bool) {
	if queue[i].used {
		return nil, false
	}
	buf = append(buf, queue[i])
	for j := i + 1; j < len(queue); j++ {
		if !queue[j].used && queue[j].matchTeam(buf) {
			queue[j].used = true
			buf = append(buf, queue[j])
			if len(buf) == n {
				return buf, true
			}
		}
	}
	for _, t := range buf {
		t.used = false
	}
	return nil, false
}

func removeUsed(queue [][]*Ticket) {
	for i := 0; i < len(queue); i++ {
		queue[i] = slices.DeleteFunc(queue[i], func(t *Ticket) bool {
			return t == nil || t.used
		})
	}
}

func search(buf *candidate, queue [][]*Ticket, need int, cut, hate bool) (ok bool) {
	var (
		x        int
		due2Hate bool
	)

SEARCH:
	for buf.high < need {
		if cut {
			x = min(need, need-buf.low)
		} else {
			x = min(need, need-buf.high)
		}
		for i := x; i >= 1; i-- {
			for j := range queue[i] {
				if !queue[i][j].used {
					if !hate || buf.allow(queue[i][j]) {
						buf.join(queue[i][j])
						continue SEARCH
					}
					due2Hate = true
				}
			}
		}
		break
	}
	if buf.low <= need && need <= buf.high {
		return true
	}
	if !due2Hate {
		for _, t := range buf.tickets {
			t.used = false
		}
	}
	return false
}

func quickFail(queue [][]*Ticket, max, team int) bool {
	sum := 0
	for i := 1; i <= max; i++ {
		sum += len(queue[i]) * i
	}
	return sum/max < team
}

func sortTicket(ts map[string]*Ticket, max int) [][]*Ticket {
	result := make([][]*Ticket, max+1)
	for _, t := range ts {
		n := len(t.Members)
		if n > max || n <= 0 { // ignore wrong input
			continue
		}
		result[n] = append(result[n], t)
	}
	for i := 1; i <= max; i++ {
		sort.Slice(result[i], func(a, b int) bool {
			return result[i][a].startMatch < result[i][b].startMatch
		})
	}
	return result
}
