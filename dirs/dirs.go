package dirs

import (
	"errors"
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
		p := f(app)
		if _, err := os.Stat(p); err != nil && errors.Is(err, os.ErrNotExist) {
			return p, os.MkdirAll(p, 0755)
		}
	}
	return "", ErrNotMade
}

// AppDirFindOrMake will search for app directory in the given order (or default order if none is given)
// or create it if it doesn't exist.
func AppDirFindOrMake(app string, appDirSearchOrder, appDirMakeOrder []func(string) string) (string, error) {
	ad, err := FindAppDir(app, appDirSearchOrder...)
	if err != nil {
		return MakeAppDir(app, appDirMakeOrder...)
	}
	return ad, nil
}

// MustAppDir do AppDirFindOrMake and panic if error.
func MustAppDir(app string, searchOrder ...func(string) string) string {
	appDir, err := AppDirFindOrMake(app, searchOrder, searchOrder)
	if err != nil {
		panic(err)
	}
	return appDir
}

// AppFile will return a path to a file in the app directory. Creating the app directory if it doesn't exist.
func AppFile(app, fileName string, searchOrder ...func(string) string) (string, error) {
	appFile, err := AppDirFindOrMake(app, searchOrder, searchOrder)
	if err != nil {
		return "", err
	}
	return path.Join(appFile, fileName), nil
}

// MustAppFile do AppFile and panic if error.
func MustAppFile(app, fil string, searchOrder ...func(string) string) string {
	appFile, err := AppFile(app, fil, searchOrder...)
	if err != nil {
		panic(err)
	}
	return appFile
}
