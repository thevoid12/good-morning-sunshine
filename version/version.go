package version

import "fmt"

const GmsMajor = 1
const GmsMinor = 1
const GmsPatch = 2
const ReleaseDate = "30-Oct-2024"

func GetLatestVersion() string {
	return "Version: " + fmt.Sprint(GmsMajor) + "." + fmt.Sprint(GmsMinor) + "." + fmt.Sprint(GmsPatch) + " " + ReleaseDate
}
