package account

// SMTPConfig SMTP配置值对象
type SMTPConfig struct {
	host     string
	port     int
	username string
	password string
	useTLS   bool
}

// NewSMTPConfig 创建SMTP配置
func NewSMTPConfig(host string, port int, username string, password string, useTLS bool) SMTPConfig {
	return SMTPConfig{
		host:     host,
		port:     port,
		username: username,
		password: password,
		useTLS:   useTLS,
	}
}

// Getters
func (c SMTPConfig) Host() string     { return c.host }
func (c SMTPConfig) Port() int        { return c.port }
func (c SMTPConfig) Username() string { return c.username }
func (c SMTPConfig) Password() string { return c.password }
func (c SMTPConfig) UseTLS() bool     { return c.useTLS }

// Address 返回服务器地址
func (c SMTPConfig) Address() string {
	return c.host + ":" + string(rune(c.port))
}

// HasAuth 是否需要认证
func (c SMTPConfig) HasAuth() bool {
	return c.username != "" && c.password != ""
}