package config

import (
	"log"
	"text/template"

	"github.com/alexedwards/scs/v2"
	"github.com/k3forx/booking-app/internal/models"
)

// AppConfig holds the application config
type AppConfig struct {
	UseCache      bool // false is development mode as cache is reloaded every time
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProduction  bool
	Session       *scs.SessionManager
	MailChan      chan models.MailData
}
