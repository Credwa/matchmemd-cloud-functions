package main

import (
	"context"
	"log"
	"os"

	matchmemdpasswordreset "matchmemd-password-reset"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"matchmemd.com/functions/common"
)

func main() {
	ctx := context.Background()

	if err := funcframework.RegisterHTTPFunctionContext(ctx, "/", common.DefaultRequest); err != nil {
		log.Fatalf("funcframework.RegisterHTTPFunctionContext: %v\n", err)
	}

	if err := funcframework.RegisterHTTPFunctionContext(ctx, "/password-reset-request", matchmemdpasswordreset.PasswordResetRequest); err != nil {
		log.Fatalf("funcframework.RegisterHTTPFunctionContext: %v\n", err)
	}
	port := "8765"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	log.Printf("Serving on port: %s \n", port)
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
