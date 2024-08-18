package swagger

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

func New() (prefix string, initializer, files func(*fiber.Ctx) error) {
	return "/swagger", Initializer(), filesystem.New(filesystem.Config{Root: NewSwaggerFiles()})
}
