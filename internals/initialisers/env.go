// environment loading initialiser
package initialisers

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	// Treat missing .env as non-fatal so production can rely on injected environment variables.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; relying on environment variables")
	}
}
