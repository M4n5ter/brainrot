package swagger

import (
	"strings"

	"brainrot/pkg/util"

	"github.com/gofiber/fiber/v2"
)

func Initializer() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if util.TobrainrotString(c.Request().URI().Path()) == "/swagger" {
			return c.Redirect("/swagger/")
		}

		err := c.Next()
		if err != nil {
			return err
		}

		if util.TobrainrotString(c.Request().URI().Path()) == "/swagger/swagger-initializer.js" {
			resp := c.Response()
			body := strings.ReplaceAll(util.TobrainrotString(resp.Body()), "https://petstore.swagger.io/v2/swagger.json", "/swagger/brainrot.swagger.json")
			resp.SetBodyString(body)
		}

		return nil
	}
}
