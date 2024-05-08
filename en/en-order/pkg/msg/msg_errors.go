package msg

import (
	"en-order/pkg/file"
	"fmt"
	"strings"
)

func (m *MessagesErrors) doMapErrors() *MessagesErrors {
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

func SetupErrors(fname string, paths ...string) (err error) {
	var mcfg MessagesErrorsConfig
	onceErrors.Do(func() {
		msgsErrors = make(map[string]*MessagesErrors)
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

	for _, mErrors := range mcfg.Messages {
		if _, ok := msgsErrors[mErrors.Name]; !ok {
			mErrors := &MessagesErrors{Name: mErrors.Name, Contents: mErrors.Contents}
			msgsErrors[mErrors.Name] = mErrors.doMapErrors()
		}
	}
	return
}

// GetMessageErrors by language
func GetMessageErrors(key, lang string) (text string) {
	lang = cleanLangStr(lang)
	if m, ok := msgsErrors[key]; ok {
		if c, ok := m.contents[lang]; ok {
			text = c.Text
			return
		}

		return GetMessageErrors(key, DefaultLang)
	}
	return
}
