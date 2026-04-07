package email

// Flag 邮件标记枚举
type Flag string

const (
	FlagSeen     Flag = "Seen"     // 已读
	FlagFlagged  Flag = "Flagged"  // 标记/星标
	FlagDeleted  Flag = "Deleted"  // 已删除
	FlagDraft    Flag = "Draft"    // 草稿
	FlagAnswered Flag = "Answered" // 已回复
)

// Flags 标记集合值对象
type Flags []Flag

// NewFlags 创建标记集合
func NewFlags(flags ...Flag) Flags {
	return Flags(flags)
}

// Has 检查是否包含某个标记
func (f Flags) Has(flag Flag) bool {
	for _, fl := range f {
		if fl == flag {
			return true
		}
	}
	return false
}

// Add 添加标记
func (f Flags) Add(flag Flag) Flags {
	if f.Has(flag) {
		return f
	}
	return append(f, flag)
}

// Remove 移除标记
func (f Flags) Remove(flag Flag) Flags {
	result := make(Flags, 0, len(f))
	for _, fl := range f {
		if fl != flag {
			result = append(result, fl)
		}
	}
	return result
}

// IsSeen 是否已读
func (f Flags) IsSeen() bool {
	return f.Has(FlagSeen)
}

// IsFlagged 是否已标记
func (f Flags) IsFlagged() bool {
	return f.Has(FlagFlagged)
}

// IsDeleted 是否已删除
func (f Flags) IsDeleted() bool {
	return f.Has(FlagDeleted)
}

// IsDraft 是否为草稿
func (f Flags) IsDraft() bool {
	return f.Has(FlagDraft)
}

// Strings 返回字符串列表
func (f Flags) Strings() []string {
	result := make([]string, len(f))
	for i, fl := range f {
		result[i] = string(fl)
	}
	return result
}