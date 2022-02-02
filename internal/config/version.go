package config

import "strings"

var AppVersion = "unknown"
var AppTag string

func init() {
	var tag string
	tagPos := strings.Index(AppVersion, "appversion")
	if tagPos == -1 {
		return
	}

	tagPos = strings.Index(AppVersion[tagPos:], ":")
	if tagPos == -1 {
		return
	}
	tagPos++
	tagPosEnd := strings.IndexAny(AppVersion[tagPos:], "\n[")
	if tagPosEnd == -1 {
		tag = AppVersion[tagPos:]
	} else {
		tag = AppVersion[tagPos : tagPosEnd+tagPos]
	}
	AppTag = strings.TrimSpace(tag)
}
