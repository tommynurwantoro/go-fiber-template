package main

import "app/cmd"

// @title go-fiber-template API documentation
// @version 1.0.0
// @license.name MIT
// @license.url https://github.com/tommynurwantoro/go-fiber-template/blob/main/LICENSE
// @host localhost:8888
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Example Value: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
func main() {
	cmd.Execute()
}
