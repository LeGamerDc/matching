package fifo

type StringArg struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type IntArg struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

type FloatArg struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

const (
	EqualOp    = "="
	NotEqualOp = "!="
)

type StringFilter struct {
	Arg   string `json:"arg"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

type IntFilter struct {
	Arg      string  `json:"arg"`
	Min      int64   `json:"min"`
	Max      int64   `json:"max"`
	Excludes []int64 `json:"excludes"`
}

type FloatFilter struct {
	Arg string  `json:"arg"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type Ticket struct {
	TicketId   string      `json:"ticket_id"` // 整个匹配场里ticket的全局唯一id
	Members    []Member    `json:"members"`
	StringArgs []StringArg `json:"string_args"`
	IntArgs    []IntArg    `json:"int_args"`
	FloatArgs  []FloatArg  `json:"float_args"`
	BlackList  []int64     `json:"black_list"` // 设置黑名单，不会跟指定 member 匹配到同team
	startMatch int64       // 开始匹配时间，epoch 单位ms
	endMatch   int64       // 结束匹配时间，epoch 单位ms
	used       bool
}

type Member struct {
	MemberId string `json:"member_id"` // 整个匹配匹配场里 member 的全局唯一id
	BlackId  int64  `json:"black_id"`  // 开启黑名单时，用这个值做检查
	Extra    []byte `json:"extra"`     // 用户额外数据，随匹配结果透传
	Sort     int    `json:"sort"`      // >0 表示该member是可选项，如果整个team匹配不成功可以丢弃该member匹配
}

type MatchResult struct {
	PoolName string       `json:"pool_name"` // 表示从哪个池子挑出来的
	Teams    []TeamResult `json:"teams"`     // 挑选出来的队伍
}

type TeamResult struct {
	TeamName   string         `json:"team_name"` // 表示对应哪个 TeamProfile
	TicketId   []string       `json:"ticket_id"` // 表示来源于哪些 Ticket
	Members    []MemberResult `json:"members"`
	CutMembers []MemberResult `json:"cut_members"`
}

type MemberResult struct {
	MemberId string `json:"member_id"` // 表示来源于Team中哪个 Member
	Extra    []byte `json:"extra"`     // 从 Member 携带的 Extra
	sort     int
}

// MatchProfile 匹配场配置
type MatchProfile struct {
	Name      string        `json:"name"`      // 匹配场名字，唯一
	Algorithm string        `json:"algorithm"` // 进行匹配所使用的算法，内置 fifo
	Tick      string        `json:"tick"`      // 匹配场匹配频率，如 "0.5s"
	Pools     []PoolProfile `json:"pools"`     // 匹配池
}

// PoolProfile 匹配池配置
type PoolProfile struct {
	Name                    string         `json:"name"`                       // 匹配池名字
	BetweenTeamAntiAffinity string         `json:"between_team_anti_affinity"` // 队间反亲和性选择词
	StringFilters           []StringFilter `json:"string_filters"`
	IntFilters              []IntFilter    `json:"int_filters"`
	FloatFilters            []FloatFilter  `json:"float_filters"`
	Teams                   []string       `json:"teams"`               // 匹配结果需要多个team
	TeamMembers             int            `json:"team_members"`        // 每队人数
	MaxMatchPerRound        int            `json:"max_match_per_round"` // 每场最多匹配队伍
	AllowCut                bool           `json:"allow_cut"`           // 允许缩减队伍
	AllowHate               bool           `json:"allow_hate"`          // 考虑玩家的黑名单
}

type ResultSubmitter func(MatchResult)
