package chelper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func ReadJsonFile(path string, out interface{}) error {
	data, err := ioutil.ReadFile(path)
	if nil != err {
		return err
	}
	return json.Unmarshal(data, out)
}

func NormalizePath(p string) string {
	if "" == p {
		return p
	}
	p = path.Clean(p)
	p = strings.Replace(p, "\\", "/", -1)
	p = strings.Replace(p, "//", "/", -1)
	return p
}

func BinPathSplit() (dir, name, ext string, err error) {
	var binPath string
	binPath, err = os.Executable()
	if nil != err {
		err = errors.New(fmt.Sprintf("get bin path fail: %v", err))
		return
	}
	binPath, err = filepath.Abs(binPath)
	if nil != err {
		err = errors.New(fmt.Sprintf("get abs path fail: %v", err))
		return
	}
	dir, name, ext = PathSplit(binPath)
	return
}

func PathSplit(fullpath string) (dir, name, ext string) {
	fullpath = NormalizePath(fullpath)
	dir, name = path.Split(fullpath)
	idx := strings.LastIndex(name, ".")
	if 0 < idx {
		ext = name[idx:]
		name = name[:idx]
	}
	return
}
