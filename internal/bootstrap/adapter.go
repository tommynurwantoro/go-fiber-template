package bootstrap

import (
	"app/internal/adapter/database"
	"app/internal/adapter/email"
	"app/internal/adapter/oauth"
	"app/internal/adapter/rest"
)

func RegisterAdapters() {
	appContainer.RegisterService("database", new(database.Gorm))
	appContainer.RegisterService("rest", new(rest.Fiber))
	appContainer.RegisterService("email", new(email.EmailAdapterImpl))
	appContainer.RegisterService("oauth", new(oauth.GoogleAdapterImpl))
}
