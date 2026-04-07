package email

// Body 邮件正文值对象
type Body struct {
	text    string
	html    string
	charset string
}

// NewBody 创建新的正文
func NewBody(text string, html string, charset string) Body {
	return Body{
		text:    text,
		html:    html,
		charset: charset,
	}
}

// NewTextBody 创建纯文本正文
func NewTextBody(text string) Body {
	return Body{
		text:    text,
		charset: "utf-8",
	}
}

// NewHTMLBody 创建HTML正文
func NewHTMLBody(html string) Body {
	return Body{
		html:    html,
		charset: "utf-8",
	}
}

// Getters
func (b Body) Text() string    { return b.text }
func (b Body) HTML() string    { return b.html }
func (b Body) Charset() string { return b.charset }

// HasText 是否有纯文本内容
func (b Body) HasText() bool {
	return b.text != ""
}

// HasHTML 是否有HTML内容
func (b Body) HasHTML() bool {
	return b.html != ""
}

// IsEmpty 是否为空
func (b Body) IsEmpty() bool {
	return b.text == "" && b.html == ""
}

// PreferredBody 返回优先显示的正文（优先HTML）
func (b Body) PreferredBody() string {
	if b.HasHTML() {
		return b.html
	}
	return b.text
}

// PlainText 返回纯文本版本（如果没有则从HTML提取）
func (b Body) PlainText() string {
	if b.HasText() {
		return b.text
	}
	// 如果只有HTML，返回HTML（实际使用时可能需要进一步处理）
	return b.html
}