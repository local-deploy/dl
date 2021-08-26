package lang

import (
	"github.com/bykovme/gotrans"
)

func Text(key string) string {
	_ = gotrans.InitLocales("./lang")

	return gotrans.T(key)
}

func SetLang(locale string) {
	_ = gotrans.SetDefaultLocale(locale)
}
