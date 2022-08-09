package helper

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// HomeDir user home directory
func HomeDir() (string, error) {
	return os.UserHomeDir()
}

// ConfigDir config directory (~/.config/dl)
func ConfigDir() (string, error) {
	home, err := HomeDir()

	return filepath.Join(home, ".config", "dl"), err
}

// BinDir path to local bin directory
func BinDir() (string, error) {
	home, err := HomeDir()

	return filepath.Join(home, ".local", "bin"), err
}

// BinPath path to bin
func BinPath() (string, error) {
	binDir, err := BinDir()

	return filepath.Join(binDir, "dl"), err
}

// IsConfigDirExists checking for the existence of a configuration directory
func IsConfigDirExists() bool {
	confDir, _ := ConfigDir()
	_, err := os.Stat(confDir)

	return err == nil
}

// IsConfigFileExists checking for the existence of a configuration file
func IsConfigFileExists() bool {
	confDir, _ := ConfigDir()
	config := filepath.Join(confDir, "config.yaml")

	_, err := os.Stat(config)

	return err == nil
}

// IsBinFileExists checks the existence of a binary
func IsBinFileExists() bool {
	binDir, _ := BinDir()
	bin := filepath.Join(binDir, "dl")

	_, err := os.Stat(bin)

	return err == nil
}

// ChmodR change file permissions recursively
func ChmodR(path string, mode os.FileMode) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chmod(name, mode)
		}

		return err
	})
}

// CallMethod is necessary to avoid map of functions
func CallMethod(i interface{}, methodName string) interface{} {
	var ptr reflect.Value
	var value reflect.Value
	var finalMethod reflect.Value

	value = reflect.ValueOf(i)

	if value.Type().Kind() == reflect.Ptr {
		ptr = value
		value = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(i))
		temp := ptr.Elem()
		temp.Set(value)
	}

	// check for method on value
	method := value.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}
	// check for method on pointer
	method = ptr.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}

	if finalMethod.IsValid() {
		return finalMethod.Call([]reflect.Value{})[0].Interface()
	}

	i = make([]string, 0)
	return i
}

func ContainsHasPrefix(s []string, e string) bool {

	for _, a := range s {
		if strings.HasPrefix(e, a) || a == e {
			return true
		}
	}
	return false
}