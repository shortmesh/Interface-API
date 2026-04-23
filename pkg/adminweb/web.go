package adminweb

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed web/dist
var distFS embed.FS

func RegisterRoutes(e *echo.Echo) error {
	distSubFS, err := fs.Sub(distFS, "web/dist")
	if err != nil {
		return err
	}

	e.GET("/admin", echo.WrapHandler(http.StripPrefix("/admin", http.FileServer(http.FS(distSubFS)))))
	e.GET("/admin/*", echo.WrapHandler(http.StripPrefix("/admin", http.FileServer(http.FS(distSubFS)))))

	return nil
}
