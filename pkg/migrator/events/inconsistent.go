package events

// InconsistentEvent 数据不一致，上报给kafka修复的事件定义
type InconsistentEvent struct {
	ID        int64
	Direction string // 用什么来修，取值为 SRC，意味着，以源表为准，取值为 DST，以目标表为准
	Type      string
}

const (
	// InconsistentEventTypeTargetMissing 校验的目标数据，缺了这一条
	InconsistentEventTypeTargetMissing = "target_missing"
	// InconsistentEventTypeNEQ 数据不相等
	InconsistentEventTypeNEQ = "neq"
	// InconsistentEventTypeBaseMissing 校验的目标数据，多了这一条
	InconsistentEventTypeBaseMissing = "base_missing"
)
