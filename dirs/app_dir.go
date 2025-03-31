package dirs

import (
	"errors"
	"fmt"
	"os"
	"path"
)

func NewAppDir(app string, searchOrder ...func(string) string) (AppDir, error) {
	appDir, err := AppDirFindOrMake(app, searchOrder, searchOrder)
	return AppDir{
		appDir: appDir,
	}, err
}

type AppDir struct {
	appDir string
}

func (a AppDir) String() string {
	return a.appDir
}

func (a AppDir) dir(create bool, name ...string) (string, error) {
	if len(name) == 0 {
		return "", fmt.Errorf("no dir name given")
	}
	d := path.Join(append([]string{a.appDir}, name...)...)
	fi, err := os.Stat(d)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if !create {
				return d, err
			}
			return d, os.MkdirAll(d, 0755)
		}
		return "", fmt.Errorf("unable to determine dir's (%s) existence: %s", d, err)
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("not a directory: %s", d)
	}
	return d, nil
}

// Dir will return a path to a new in the app directory.
// It will create the dir.
func (a AppDir) Dir(name ...string) (string, error) {
	return a.dir(true, name...)
}

// MightDir will return a path to a directory in the app directory.
// It will NOT create the dir but WILL return the path or panic if it's not possible.
func (a AppDir) MightDir(name ...string) string {
	d, err := a.dir(false, name...)
	if errors.Is(err, os.ErrNotExist) {
		return d
	}
	if err != nil {
		panic(err)
	}
	return d
}

// MustDir will return a path to a new in the app directory.
// It WILL create the dir or panic if it's not possible.
func (a AppDir) MustDir(name ...string) string {
	d, err := a.Dir(name...)
	if err != nil {
		panic(err)
	}
	return d
}

// File will return a path to a file in the app directory,
// creating the app directory (and all required parent directories) if it doesn't exist.
func (a AppDir) File(name ...string) (string, error) {
	if len(name) == 0 {
		return "", fmt.Errorf("no file name given")
	}
	if len(name) > 1 {
		_, err := a.Dir(name[:len(name)-1]...)
		if err != nil {
			return "", err
		}
	}
	f := path.Join(append([]string{a.appDir}, name...)...)
	fInfo, err := os.Stat(f)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return f, nil
		}
		return "", fmt.Errorf("unable to determine file's (%s) existence: %s", f, err)
	}
	if fInfo.IsDir() {
		return "", fmt.Errorf("not a file: %s", f)
	}
	return f, nil
}

func (a AppDir) MustFile(name ...string) string {
	f, err := a.File(name...)
	if err != nil {
		panic(err)
	}
	return f
}
