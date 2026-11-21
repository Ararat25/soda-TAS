package main

import (
	"log"
	"os"

	"github.com/Ararat25/soda-TAS/internal/container"
	"github.com/Ararat25/soda-TAS/internal/transport/cli"
)

func main() {
	cont, err := container.New()
	if err != nil {
		log.Fatal(err)
	}

	err = cont.Invoke(func(app *cli.CLI) error {
		return app.Run(os.Args[1:])
	})

	if err != nil {
		log.Fatal(err)
	}
}
