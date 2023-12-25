package httpserver

import (
	"awesomeProject1/config"
	"awesomeProject1/db/model"
	"awesomeProject1/db/postgres"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/template/html/v2"
	"os"
)

func RunServer() {
	cfg := config.MustParseConfig()
	db, err := postgres.NewPostgresDatabase(cfg.PostgresDSN)
	if err != nil {
		fmt.Printf("Error connecting to the database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	engine := html.New("./apps/httpserver/templates", ".gohtml")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Static("/static", "./apps/imagecrawler/"+cfg.ImagesFolderName)
	app.Get("/", func(c *fiber.Ctx) error {
		allImages, err := db.ListImages(c.UserContext(), &model.ListImagesFilters{})
		if err != nil {
			return err
		}
		filteredImages, err := db.ListImages(c.UserContext(), &model.ListImagesFilters{
			Format: model.ImageFormat(c.Query("format", "")),
		})
		if err != nil {
			return err
		}

		return c.Render("search", fiber.Map{
			"Images":  filteredImages,
			"Formats": model.ImageFormatValues(),
			"Counts":  model.ImagesCountPerFormat(allImages),
		})
	})
	app.Get("/images/:filename", func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		return c.SendFile("/static/" + filename)
	})

	log.Fatal(app.Listen(":8080"))
}
