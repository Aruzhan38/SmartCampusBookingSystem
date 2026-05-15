package http

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*.html
var templateFS embed.FS

func loadTemplates() *template.Template {
	tmpl := template.Must(template.New("").ParseFS(templateFS, "templates/*.html"))
	return tmpl
}

type WebHandler struct {
}

func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

func (h *WebHandler) Templates() *template.Template {
	return loadTemplates()
}

func (h *WebHandler) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func (h *WebHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func (h *WebHandler) RegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}

func (h *WebHandler) RoomsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "rooms.html", nil)
}

func (h *WebHandler) BookingsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "bookings.html", nil)
}

func (h *WebHandler) ProfilePage(c *gin.Context) {
	c.HTML(http.StatusOK, "profile.html", nil)
}
