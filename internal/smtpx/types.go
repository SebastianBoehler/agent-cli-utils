package smtpx

import "time"

type Config struct {
	Provider     string        `json:"provider,omitempty" yaml:"provider,omitempty"`
	Host         string        `json:"host,omitempty" yaml:"host,omitempty"`
	Port         int           `json:"port,omitempty" yaml:"port,omitempty"`
	Security     string        `json:"security,omitempty" yaml:"security,omitempty"`
	Auth         string        `json:"auth,omitempty" yaml:"auth,omitempty"`
	Username     string        `json:"username,omitempty" yaml:"username,omitempty"`
	Password     string        `json:"password,omitempty" yaml:"password,omitempty"`
	PasswordEnv  string        `json:"password_env,omitempty" yaml:"password_env,omitempty"`
	PasswordFile string        `json:"password_file,omitempty" yaml:"password_file,omitempty"`
	From         string        `json:"from,omitempty" yaml:"from,omitempty"`
	Timeout      time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

type Profile struct {
	Operation    string `json:"operation" yaml:"operation"`
	Provider     string `json:"provider" yaml:"provider"`
	Host         string `json:"host" yaml:"host"`
	Port         int    `json:"port" yaml:"port"`
	Security     string `json:"security" yaml:"security"`
	Auth         string `json:"auth" yaml:"auth"`
	Username     string `json:"username,omitempty" yaml:"username,omitempty"`
	From         string `json:"from,omitempty" yaml:"from,omitempty"`
	Timeout      string `json:"timeout" yaml:"timeout"`
	HasSecret    bool   `json:"has_secret" yaml:"has_secret"`
	SecretSource string `json:"secret_source,omitempty" yaml:"secret_source,omitempty"`
}

type Message struct {
	To      []string `json:"to" yaml:"to"`
	Cc      []string `json:"cc,omitempty" yaml:"cc,omitempty"`
	Bcc     []string `json:"bcc,omitempty" yaml:"bcc,omitempty"`
	Subject string   `json:"subject,omitempty" yaml:"subject,omitempty"`
	Text    string   `json:"text,omitempty" yaml:"text,omitempty"`
}

type Result struct {
	Operation    string   `json:"operation" yaml:"operation"`
	Status       string   `json:"status" yaml:"status"`
	Provider     string   `json:"provider" yaml:"provider"`
	Host         string   `json:"host" yaml:"host"`
	Port         int      `json:"port" yaml:"port"`
	Security     string   `json:"security" yaml:"security"`
	Auth         string   `json:"auth" yaml:"auth"`
	Username     string   `json:"username,omitempty" yaml:"username,omitempty"`
	From         string   `json:"from,omitempty" yaml:"from,omitempty"`
	To           []string `json:"to,omitempty" yaml:"to,omitempty"`
	Cc           []string `json:"cc,omitempty" yaml:"cc,omitempty"`
	Bcc          []string `json:"bcc,omitempty" yaml:"bcc,omitempty"`
	Subject      string   `json:"subject,omitempty" yaml:"subject,omitempty"`
	MessageBytes int      `json:"message_bytes,omitempty" yaml:"message_bytes,omitempty"`
}
