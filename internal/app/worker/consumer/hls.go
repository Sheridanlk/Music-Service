package consumer

import (
	"context"
	"log/slog"
	"strconv"
	"sync"

	"github.com/Sheridanlk/Music-Service/internal/services/tracks/hls"
	"github.com/rabbitmq/amqp091-go"
)

type HlsConsumer struct {
	log        *slog.Logger
	hlsService *hls.HlsSegmenter
	messages   <-chan amqp091.Delivery
}

func New(log *slog.Logger, hlsService *hls.HlsSegmenter, messages <-chan amqp091.Delivery) *HlsConsumer {
	return &HlsConsumer{
		log:        log,
		hlsService: hlsService,
		messages:   messages,
	}
}

func (h *HlsConsumer) Consume(ctx context.Context) <-chan struct{} {
	op := "HlsConsumer.Consume"

	log := h.log.With(
		slog.String("op", op),
	)

	done := make(chan struct{})

	h.log.Info("hls consumer started")

	go func() {
		defer close(done)

		var wg sync.WaitGroup

		for {
			select {
			case <-ctx.Done():
				log.Info("hls consumer stopped")
				wg.Wait()
				return

			case d, ok := <-h.messages:
				if !ok {
					log.Info("hls consumer stopped, channel closed")
					return
				}

				wg.Add(1)

				go func(msg amqp091.Delivery) {
					defer wg.Done()

					idStr := string(msg.Body)
					id, err := strconv.ParseInt(idStr, 10, 64)
					if err != nil {
						log.Error("failed to parse track ID", "error", err)

						_ = msg.Nack(false, false)

						return
					}

					log.Info("processing track for hls segmentation", "track_id", id)

					if err := h.hlsService.Hls(ctx, id); err != nil {
						log.Error("failed to segment track for hls", "track_id", id, "error", err)

						_ = msg.Nack(false, true)

						return
					}

					_ = msg.Ack(false)

					log.Info("finished processing track for hls segmentation", "track_id", id)
				}(d)
			}
		}
	}()

	return done
}
