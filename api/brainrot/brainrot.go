package main

import (
	"context"
	"log"

	"brainrot/api/brainrot/gatewayoption"
	"brainrot/api/brainrot/middleware/swagger"
	"brainrot/gen/pb/brainrot"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	gwmux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(gatewayoption.HeaderMatcher),
		runtime.WithMarshalerOption(gatewayoption.NewResponseWrapper()),
		runtime.WithErrorHandler(gatewayoption.ErrorHandler),
		// runtime.WithForwardResponseOption(ForwardResponse), // 当 grpc 返回错误时，不会触发 ForwardResponse，而是提前走 runtime.HTTPError 后返回
		runtime.WithMetadata(gatewayoption.WithMetadata),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := brainrot.RegisterPingHandlerFromEndpoint(context.Background(), gwmux, ":8080", opts)
	if err != nil {
		log.Println(err)
		return
	}

	err = brainrot.RegisterUserHandlerFromEndpoint(context.Background(), gwmux, ":8080", opts)
	if err != nil {
		log.Println(err)
		return
	}

	err = brainrot.RegisterArticleHandlerFromEndpoint(context.Background(), gwmux, ":8080", opts)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Starting gateway server at %s...\n", "8090")
	gwApp := fiber.New()

	gwApp.Use(compress.New())
	gwApp.Use(helmet.New())
	gwApp.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173", // 允许的前端地址
		AllowCredentials: true,
	}))
	gwApp.Use(idempotency.New())
	gwApp.Use(etag.New())
	gwApp.Use(swagger.New())

	gwApp.All("/*", adaptor.HTTPHandler(gwmux))

	if err := gwApp.Listen(":8090"); err != nil {
		log.Println(err)
		return
	}
}
