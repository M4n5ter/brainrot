package swagger

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/m4n5ter/makabaka/pkg/util"
)

func Initializer() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if util.ToMakabakaString(c.Request().URI().Path()) == "/swagger" {
			return c.Redirect("/swagger/")
		}

		err := c.Next()
		if err != nil {
			return err
		}

		if util.ToMakabakaString(c.Request().URI().Path()) == "/swagger/swagger-initializer.js" {
			resp := c.Response()
			body := strings.ReplaceAll(util.ToMakabakaString(resp.Body()), "https://petstore.swagger.io/v2/swagger.json", "/swagger/makabaka.swagger.json")
			resp.SetBodyString(body)
		}

		return nil
	}
}
