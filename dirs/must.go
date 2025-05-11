package dirs

// MustAppDir do AppDirFindOrMake and panic if error.
func MustAppDir(app string, searchOrder ...func(string) string) string {
	appDir, err := AppDirFindOrMake(app, searchOrder, searchOrder)
	if err != nil {
		panic(err)
	}
	return appDir
}

// MustAppFile do AppFile and panic if error.
func MustAppFile(app, fil string, searchOrder ...func(string) string) string {
	appFile, err := AppFile(app, fil, searchOrder...)
	if err != nil {
		panic(err)
	}
	return appFile
}
