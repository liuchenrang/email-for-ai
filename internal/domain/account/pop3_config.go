package account

// POP3Config POP3配置值对象
type POP3Config struct {
	host     string
	port     int
	username string
	password string
	useTLS   bool
}

// NewPOP3Config 创建POP3配置
func NewPOP3Config(host string, port int, username string, password string, useTLS bool) POP3Config {
	return POP3Config{
		host:     host,
		port:     port,
		username: username,
		password: password,
		useTLS:   useTLS,
	}
}

// Getters
func (c POP3Config) Host() string     { return c.host }
func (c POP3Config) Port() int        { return c.port }
func (c POP3Config) Username() string { return c.username }
func (c POP3Config) Password() string { return c.password }
func (c POP3Config) UseTLS() bool     { return c.useTLS }

// Address 返回服务器地址
func (c POP3Config) Address() string {
	return c.host + ":" + string(rune(c.port))
}

// HasAuth 是否需要认证
func (c POP3Config) HasAuth() bool {
	return c.username != "" && c.password != ""
}