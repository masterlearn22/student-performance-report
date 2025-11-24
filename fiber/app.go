package fiber
import (
	"github.com/gofiber/fiber/v2"
	"fmt"
	
	
)

func SetupFiber() *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			
			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
				
			})
		},
	
	})
	fmt.Println("Fiber is running")
	return app
}