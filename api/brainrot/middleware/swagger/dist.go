package swagger

import (
	"embed"
	"io/fs"
	"net/http"

	swaggerfiles "github.com/swaggo/files/v2"
)

type Files struct {
	dist http.FileSystem
	json http.FileSystem
}

func NewSwaggerFiles() Files {
	json, _ := fs.Sub(swaggerjson, "proto")
	return Files{
		dist: http.FS(swaggerfiles.FS),
		json: http.FS(json),
	}
}

func (s Files) Open(name string) (http.File, error) {
	file, err := s.dist.Open(name)
	if err == nil {
		return file, nil
	}
	return s.json.Open(name)
}

//go:embed proto/brainrot.swagger.json
var swaggerjson embed.FS
