package lang

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

func trans() {
	lang, _ := language.Parse("ru")
	_ = message.Set(language.Russian, "Hello {name}", catalog.String("Привет %s"))
	hello := message.NewPrinter(lang).Sprintf("Hello {name}", "Andrey")
	panic(hello)
	//Output: Привет Andrey
}
