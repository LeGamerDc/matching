package fifo

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var candidates = []*Ticket{
	{
		TicketId: "1",
		Members: []Member{{
			MemberId: "1_1", BlackId: 10001,
		}},
	},
	{
		TicketId:   "2",
		StringArgs: []StringArg{{"as", "10002"}},
		Members: []Member{{
			MemberId: "2_1", BlackId: 20001,
		}, {
			MemberId: "2_2", BlackId: 20002, Sort: 1,
		}},
	},
	{
		TicketId:   "3",
		StringArgs: []StringArg{{"as", "10002"}},
		Members: []Member{{
			MemberId: "3_1", BlackId: 30001,
		}, {
			MemberId: "3_2", BlackId: 30002, Sort: 1,
		}, {
			MemberId: "3_3", BlackId: 30003, Sort: 2,
		}},
	},
	{
		TicketId:   "4",
		StringArgs: []StringArg{{"as", "10001"}},
		Members: []Member{{
			MemberId: "4_1", BlackId: 40001,
		}, {
			MemberId: "4_2", BlackId: 40002, Sort: 1,
		}, {
			MemberId: "4_3", BlackId: 40003, Sort: 2,
		}, {
			MemberId: "4_4", BlackId: 40004, Sort: 3,
		}},
	},
	{
		TicketId:   "5",
		StringArgs: []StringArg{{"as", "10001"}},
		Members: []Member{{
			MemberId: "5_1", BlackId: 50001,
		}, {
			MemberId: "5_2", BlackId: 50002, Sort: 1,
		}, {
			MemberId: "5_3", BlackId: 50003, Sort: 2,
		}, {
			MemberId: "5_4", BlackId: 50004, Sort: 3,
		}, {
			MemberId: "5_5", BlackId: 50005, Sort: 4,
		}},
	},
}

func Test_quickFail(t *testing.T) {
	q := make([][]*Ticket, 6)
	q[5] = make([]*Ticket, 2)
	assert.False(t, quickFail(q, 5, 2))

	q[5] = nil
	q[4] = make([]*Ticket, 1)
	q[3] = make([]*Ticket, 1)
	q[1] = make([]*Ticket, 2)
	assert.True(t, quickFail(q, 5, 2))

	q[1] = make([]*Ticket, 3)
	assert.False(t, quickFail(q, 5, 2))
}

func Test_sortTicket(t *testing.T) {
	tickets := make(map[string]*Ticket)
	for _, ti := range candidates {
		tickets[ti.TicketId] = ti
	}
	q := sortTicket(tickets, 5)
	for i := 1; i <= 5; i++ {
		assert.Equal(t, 1, len(q[i]))
	}
}

func Test_FifoMatch(t *testing.T) {
	tickets := make(map[string]*Ticket)
	for _, ti := range candidates {
		tickets[ti.TicketId] = ti
	}
	var results []MatchResult
	FifoMatch(PoolProfile{
		Teams:            []string{"1"},
		TeamMembers:      5,
		MaxMatchPerRound: 100,
		AllowCut:         false,
	}, tickets, time.Now().UnixMilli(), func(result MatchResult) {
		results = append(results, result)
	})
	fmt.Printf("%+v\n", results)
}

func Test_FifoMatch2(t *testing.T) {
	tickets := make(map[string]*Ticket)
	tickets["3"] = candidates[2]
	tickets["4"] = candidates[3]
	var results []MatchResult
	FifoMatch(PoolProfile{
		Teams:            []string{"1"},
		TeamMembers:      5,
		MaxMatchPerRound: 100,
		AllowCut:         true,
	}, tickets, time.Now().UnixMilli(), func(result MatchResult) {
		results = append(results, result)
	})
	fmt.Printf("%+v\n", results)
}

func Test_FifoMatch3(t *testing.T) {
	tickets := make(map[string]*Ticket)
	for _, ti := range candidates {
		tickets[ti.TicketId] = ti
	}
	var results []MatchResult
	FifoMatch(PoolProfile{
		Teams:                   []string{"1", "2"},
		TeamMembers:             5,
		MaxMatchPerRound:        100,
		BetweenTeamAntiAffinity: "as",
	}, tickets, time.Now().UnixMilli(), func(result MatchResult) {
		results = append(results, result)
	})
	fmt.Printf("%+v\n", results)
}
