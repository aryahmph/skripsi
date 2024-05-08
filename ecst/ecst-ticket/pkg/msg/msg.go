package msg

import (
	"ecst-ticket/pkg/file"
	"fmt"
	"net/http"
	"strings"
)

func (m *Messages) doMap() *Messages {
	m.contents = make(map[string]*Contents)
	for _, c := range m.Contents {
		l := strings.ToLower(c.Lang)
		if _, ok := m.contents[l]; !ok {
			m.contents[l] = c
			continue
		}
	}
	return m
}

func Setup(fname string, paths ...string) (err error) {
	var mcfg MessageConfig
	once.Do(func() {
		msgs = make(map[string]*Messages)
		for _, p := range paths {
			f := fmt.Sprint(p, fname)
			err = file.ReadFromYAML(f, &mcfg)
			if err != nil {
				continue
			}
			err = nil
		}
	})

	if err != nil {
		err = fmt.Errorf("unable to read config from files %s", err.Error())
		return
	}

	for _, m := range mcfg.Messages {
		if _, ok := msgs[m.Name]; !ok {
			m := &Messages{Name: m.Name, Code: m.Code, Contents: m.Contents}
			msgs[m.Name] = m.doMap()
		}
	}
	return
}

// Get messages by language
func Get(key, lang string) (text string) {
	lang = cleanLangStr(lang)
	if m, ok := msgs[key]; ok {
		if c, ok := m.contents[lang]; ok {
			text = c.Text
			return
		}

		return Get(key, DefaultLang)
	}
	return
}

// GetCode messages by language
func GetCode(key string) int {
	if m, ok := msgs[key]; ok {
		return m.Code
	}
	return http.StatusUnprocessableEntity
}

// GetMessageCode messages by language
func GetMessageCode(key, lang string) (code int, text string) {
	lang = cleanLangStr(lang)
	if m, ok := msgs[key]; ok {
		code = m.Code
		if c, ok := m.contents[lang]; ok {
			text = c.Text
			return
		}
	}
	code = http.StatusUnprocessableEntity
	return
}

// GetAvailableLang func check language
func GetAvailableLang(key, lang string) bool {
	if m, ok := msgs[key]; ok {
		if _, ok := m.contents[lang]; ok {
			return true
		}
	}

	return false
}
