package gdrive

import (
	"google.golang.org/api/drive/v3"
)

var ScopeFull = []string{
	drive.DriveScope,
	drive.DriveAppdataScope,
	drive.DriveAppsReadonlyScope,
	drive.DriveMeetReadonlyScope,
	drive.DriveMetadataScope,
	drive.DriveMetadataReadonlyScope,
	drive.DrivePhotosReadonlyScope,
	drive.DriveReadonlyScope,
	drive.DriveScriptsScope,
}
