package dirs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
)

var ErrNotFound = errors.New("not found")
var ErrNotMade = errors.New("not made")

// AppDirUnderWorkDirDotLocal will return a path to app directory under working directory's .local directory.
func AppDirUnderWorkDirDotLocal(app string) string {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("unable to get working directory: %s", err)
		return path.Join(".local", app)
	}
	return path.Join(wd, ".local", app)
}

// DotAppDirUnderWorkDir will return a path to app directory under working directory's .app directory.
func DotAppDirUnderWorkDir(app string) string {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("unable to get working directory: %s", err)
		return path.Join("." + app)
	}
	return path.Join(wd, "."+app)
}

// DotAppDirUnderHomeDir will return a path to app directory under user's home directory.
func DotAppDirUnderHomeDir(app string) string {
	hd, err := os.UserHomeDir()
	if err != nil {
		log.Printf("unable to get user home directory: %s", err)
		return path.Join("~", "."+app)
	}
	return path.Join(hd, "."+app)
}

// AppDirUnderUserConfDir will return a path to app directory under user's config directory.
func AppDirUnderUserConfDir(app string) string {
	cd, err := os.UserConfigDir()
	if err != nil {
		log.Printf("unable to get user config directory: %s", err)
		return path.Join("~", ".config", app)
	}
	return path.Join(cd, app)
}

// AppDirUnderUserCacheDir will return a path to app directory under user's cache directory.
func AppDirUnderUserCacheDir(app string) string {
	cd, err := os.UserCacheDir()
	if err != nil {
		log.Printf("unable to get user config directory: %s", err)
		return path.Join("~", ".cache", app)
	}
	return path.Join(cd, app)
}

var defaultAppDirSearchOrder = []func(string) string{
	AppDirUnderWorkDirDotLocal,
	DotAppDirUnderWorkDir,
	DotAppDirUnderHomeDir,
	AppDirUnderUserConfDir,
	AppDirUnderUserCacheDir,
}

var defaultAppDirMakeOrder = []func(string) string{
	DotAppDirUnderWorkDir,
	DotAppDirUnderHomeDir,
	AppDirUnderUserConfDir,
	AppDirUnderUserCacheDir,
}

// FindAppDir will search for app directory in the given order (or default order if none is given).
// Defaults order is (different from make order):
// - AppDirUnderWorkDirDotLocal,
// - DotAppDirUnderWorkDir,
// - DotAppDirUnderHomeDir,
// - AppDirUnderUserConfDir,
// - AppDirUnderUserCacheDir.
func FindAppDir(app string, orderedPathFinders ...func(string) string) (string, error) {
	if len(orderedPathFinders) == 0 {
		orderedPathFinders = defaultAppDirSearchOrder
	}
	for _, f := range orderedPathFinders {
		p := f(app)
		fi, err := os.Stat(p)
		if err == nil && fi.IsDir() {
			return p, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("unable to determine dir's (%s) existence: %s", p, err)
		}
	}
	return "", ErrNotFound
}

// MakeAppDir will create app directory in the given order (or default order if none is given).
// Defaults order is (different from search order):
// - DotAppDirUnderWorkDir,
// - DotAppDirUnderHomeDir,
// - AppDirUnderUserConfDir,
// - AppDirUnderUserCacheDir.
func MakeAppDir(app string, orderedPathFinders ...func(string) string) (string, error) {
	if len(orderedPathFinders) == 0 {
		orderedPathFinders = defaultAppDirMakeOrder
	}
	for _, f := range orderedPathFinders {
		appDir := f(app)
		if _, err := os.Stat(appDir); err != nil && errors.Is(err, os.ErrNotExist) {
			return appDir, os.MkdirAll(appDir, 0755)
		}
	}
	return "", ErrNotMade
}

// AppDirFindOrMake will search for app directory in the given order (or default order if none is given)
// or create it if it doesn't exist.
func AppDirFindOrMake(app string, appDirSearchOrder, appDirMakeOrder []func(string) string) (string, error) {
	appDir, err := FindAppDir(app, appDirSearchOrder...)
	if err != nil {
		return MakeAppDir(app, appDirMakeOrder...)
	}
	return appDir, nil
}

// AppFile will return a path to a file in the app directory,
// creating the app directory if it doesn't exist.
func AppFile(app, fileName string, searchOrder ...func(string) string) (string, error) {
	appDir, err := AppDirFindOrMake(app, searchOrder, searchOrder)
	if err != nil {
		return "", err
	}
	return path.Join(appDir, fileName), nil
}

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

func (a AppDir) Dir(name string) (string, error) {
	d := path.Join(a.appDir, name)
	if fInfo, err := os.Stat(d); err == nil && fInfo.IsDir() {
		return d, nil
	}
	fi, err := os.Stat(d)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return d, os.MkdirAll(d, 0755)
		}
		return "", fmt.Errorf("unable to determine dir's (%s) existence: %s", d, err)
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("not a directory: %s", d)
	}
	return d, nil
}

func (a AppDir) MustDir(name string) string {
	d, err := a.Dir(name)
	if err != nil {
		panic(err)
	}
	return d
}
