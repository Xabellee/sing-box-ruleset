package main

import (
	sing "github.com/sagernet/sing-box/option"
)

type Config struct {
	Type    string                  `json:"type"`
	Name    string                  `json:"name"`
	Url     string                  `json:"url,omitempty"`
	Content sing.PlainRuleSetCompat `json:"content"`
}

type ConfigList struct {
	Source []Config `json:"source"`
}
