package cache

import (
	"fmt"
	"net/http"
	"time"

	"github.com/acikkaynak/backend-api-go/cache"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func New() fiber.Handler {
	cacheRepo := cache.NewRedisRepository()
	return func(c *fiber.Ctx) error {
		if c.Path() == "/healthcheck" ||
			c.Path() == "/metrics" ||
			c.Path() == "/monitor" {
			return c.Next()
		}

		reqURI := c.OriginalURL()
		hashURL := uuid.NewSHA1(uuid.NameSpaceOID, []byte(reqURI)).String()
		if c.Method() != http.MethodGet {
			// Don't cache write endpoints. We can maintain of list to exclude certain http methods later.
			// Since there will be an update in db, better to remove cache entries for this url
			err := cacheRepo.Delete(hashURL)
			if err != nil {
				fmt.Println(err)
			}
			return c.Next()
		}
		cacheData := cacheRepo.Get(hashURL)
		if cacheData == nil || len(cacheData) == 0 {
			c.Next()
			if c.Response().StatusCode() == fiber.StatusOK && len(c.Response().Body()) > 0 {
				cacheRepo.SetKey(hashURL, c.Response().Body(), 5*time.Minute)
			}
			return nil
		}

		c.Set("x-cached-response", "true")
		c.Response().SetBodyRaw(cacheData)
		c.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
		return nil
	}
}
