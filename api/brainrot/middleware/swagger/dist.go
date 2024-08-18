package swagger

import (
	"embed"
	"net/http"

	swaggerfiles "github.com/swaggo/files/v2"
)

type Files struct {
	dist http.FileSystem
	json http.FileSystem
}

func NewSwaggerFiles() Files {
	return Files{
		dist: http.FS(swaggerfiles.FS),
		json: http.FS(swaggerjson),
	}
}

func (s Files) Open(name string) (http.File, error) {
	file, err := s.dist.Open(name)
	if err == nil {
		return file, nil
	}
	return s.json.Open(name)
}

//go:embed brainrot.swagger.json
var swaggerjson embed.FS
