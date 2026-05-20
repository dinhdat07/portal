package app

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"portal-system/internal/handler/gateway"
	"portal-system/internal/handler/grpcserver"
	"syscall"
	"time"
)

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	grpcAddr := ":" + a.Config.GRPCPort
	httpAddr := ":" + a.Config.HTTPPort

	a.GRPCServer = grpcserver.NewGRPCServer(grpcserver.GRPCServerDeps{
		Validator:           a.Validator,
		Authenticator:       a.Authenticator,
		Authorizer:          a.Authorizer,
		Auth:                a.AuthGRPC,
		User:                a.UserGRPC,
		Admin:               a.AdminGRPC,
		RateLimiter:         a.RateLimiter,
		RateLimitKeyBuilder: a.RateLimitKeyBuilder,
		RateLimitConfig:     a.RateLimitConfig,
	})

	gatewayHandler, err := gateway.NewGatewayMux(ctx, grpcAddr)
	if err != nil {
		return err
	}

	a.HTTPServer = &http.Server{
		Addr:    httpAddr,
		Handler: gatewayHandler,
	}

	errCh := make(chan error, 3)

	go func() {
		log.Printf("grpc listening on %s", grpcAddr)
		errCh <- a.runGRPCServer(grpcAddr)
	}()

	go func() {
		log.Printf("gateway listening on %s", httpAddr)
		errCh <- a.runHTTPServer()
	}()

	if a.OutboxWorker != nil {
		go func() {
			log.Println("outbox worker started")
			errCh <- a.OutboxWorker.Run(ctx)
		}()
	}

	select {
	case <-ctx.Done():
		log.Println("shutdown signal received")
		return a.Shutdown(context.Background())

	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	done := make(chan struct{})

	go func() {
		if a.GRPCServer != nil {
			log.Println("stopping grpc server")
			a.GRPCServer.GracefulStop()
		}

		close(done)
	}()

	select {
	case <-done:
	case <-shutdownCtx.Done():
		if a.GRPCServer != nil {
			log.Println("grpc graceful stop timeout, forcing stop")
			a.GRPCServer.Stop()
		}
	}

	if a.HTTPServer != nil {
		log.Println("stopping gateway server")

		if err := a.HTTPServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
	}

	log.Println("application stopped gracefully")
	return nil
}

func (a *App) runHTTPServer() error {
	return a.HTTPServer.ListenAndServe()
}

func (a *App) runGRPCServer(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return a.GRPCServer.Serve(lis)
}
