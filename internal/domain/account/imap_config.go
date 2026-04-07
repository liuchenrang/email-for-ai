package account

// IMAPConfig IMAP配置值对象
type IMAPConfig struct {
	host     string
	port     int
	username string
	password string
	useTLS   bool
}

// NewIMAPConfig 创建IMAP配置
func NewIMAPConfig(host string, port int, username string, password string, useTLS bool) IMAPConfig {
	return IMAPConfig{
		host:     host,
		port:     port,
		username: username,
		password: password,
		useTLS:   useTLS,
	}
}

// Getters
func (c IMAPConfig) Host() string     { return c.host }
func (c IMAPConfig) Port() int        { return c.port }
func (c IMAPConfig) Username() string { return c.username }
func (c IMAPConfig) Password() string { return c.password }
func (c IMAPConfig) UseTLS() bool     { return c.useTLS }

// Address 返回服务器地址
func (c IMAPConfig) Address() string {
	return c.host + ":" + string(rune(c.port))
}

// HasAuth 是否需要认证
func (c IMAPConfig) HasAuth() bool {
	return c.username != "" && c.password != ""
}