package folder

import "os"

func Exist(folder string) bool {
	if info, err := os.Stat(folder); err == nil && info.IsDir() {
		return true
	}
	return false
}

func Create(path string) error {
	err := os.MkdirAll(path, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}
