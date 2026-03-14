package main

import (
	"context"

	natssever "DataConsumer/cmd/external/natsSever"
	"DataConsumer/cmd/logger"
	datastorer "DataConsumer/internal/dataStorer"
	datasubscriber "DataConsumer/internal/dataSubscriber"
	natsutil "DataConsumer/internal/natsutil"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(logger.NewLogger),
		fx.WithLogger(logger.GetFxLogger),

		fx.Supply(natsutil.NatsAddress("localhost")),
		fx.Supply(natsutil.NatsPort(4222)),
		fx.Supply(natsutil.NatsToken("")),
		fx.Supply(natsutil.NatsSeed("")),
		fx.Supply(natssever.NatsSubject("sensor.*.*.*")),
		fx.Supply(datasubscriber.BatchSize(100)),

		fx.Provide(natssever.NewNATSConnection),
		fx.Provide(natssever.NewJetStreamContext),
		fx.Provide(natssever.NewJetStreamConsumer),

		fx.Provide(datasubscriber.NewNatsDataSubscriberController),

		fx.Provide(
			fx.Annotate(
				datastorer.NewStoreDataService,
				fx.As(new(datastorer.StoreDataUseCase)),
			),
		),

		fx.Provide(func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		}),

		fx.Invoke(Init),
	).Run()
}

func Init(lc fx.Lifecycle, controller *datasubscriber.NatsDataSubscriberController, ctx context.Context, cancel context.CancelFunc, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Avvio dell'servizio DataConsumer")

			go controller.Listen()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			cancel()
			return nil
		},
	})
}
