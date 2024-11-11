package signal

import (
	"context"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func Init(parent context.Context) context.Context {
	ctx, cancel := context.WithCancel(parent)
	closeChannel := make(chan os.Signal, 1)
	signal.Notify(closeChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-closeChannel
		log.Info().Str("signal", sig.String()).Msg("Handling interruption")
		cancel()
	}()
	return ctx
}
