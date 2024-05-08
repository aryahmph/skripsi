package msg

import (
	"strings"
	"sync"
)

var (
	once        sync.Once
	onceErrors  sync.Once
	msgs        map[string]*Messages
	msgsErrors  map[string]*MessagesErrors
	DefaultLang = "id"
)

type MessageConfig struct {
	Messages []*Messages `yaml:"messages"`
}

type MessagesErrorsConfig struct {
	Messages []*MessagesErrors `yaml:"messages"`
}

type Messages struct {
	Name     string      `yaml:"name"`
	Code     int         `yaml:"code"`
	Contents []*Contents `yaml:"contents"`
	contents map[string]*Contents
}

type MessagesErrors struct {
	Name     string      `yaml:"name"`
	Contents []*Contents `yaml:"contents"`
	contents map[string]*Contents
}

type Contents struct {
	Lang string `yaml:"lang"`
	Text string `yaml:"text"`
}

type ContentsErrors struct {
	Lang string `yaml:"lang"`
	Text string `yaml:"text"`
}

func cleanLangStr(s string) string {
	return strings.ToLower(strings.Trim(s, " "))
}
