package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"kool-dev/kool/cmd"
	"kool-dev/kool/environment"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	environment.InitEnvironmentVariables(environment.NewEnvStorage(), environment.DefaultEnv)

	closingSignals()

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func closingSignals() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println("got closing signal", sig)
		os.Exit(10)
	}()
}
