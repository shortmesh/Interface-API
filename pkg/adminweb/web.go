package adminweb

import (
	"embed"
	"io/fs"
	"net/http"
	"path"

	"github.com/labstack/echo/v4"
)

//go:embed web/dist
var distFS embed.FS

func RegisterRoutes(e *echo.Echo) error {
	distSubFS, err := fs.Sub(distFS, "web/dist")
	if err != nil {
		return err
	}

	e.GET("/admin", spaHandler(distSubFS))
	e.GET("/admin/*", spaHandler(distSubFS))

	return nil
}

func spaHandler(filesystem fs.FS) echo.HandlerFunc {
	fileServer := http.FileServer(http.FS(filesystem))

	return func(c echo.Context) error {
		p := c.Request().URL.Path
		if p == "/admin" {
			p = "/"
		} else {
			p = p[len("/admin"):]
		}

		if _, err := fs.Stat(filesystem, path.Clean(p[1:])); err == nil {
			http.StripPrefix("/admin", fileServer).ServeHTTP(c.Response(), c.Request())
			return nil
		}

		c.Request().URL.Path = "/admin/"
		http.StripPrefix("/admin", fileServer).ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
