package chrono

import (
	"context"
	"log"
	"os"
	"os/signal"
	"testing"
	"time"
)

func TestNewScheduledTaskExecutor(t *testing.T) {
	scheduledTaskExecutor := NewScheduledTaskExecutor()
	scheduledTaskExecutor.ScheduleWithFixedDelay(func(ctx context.Context) {
		log.Print("hi!")
	}, 0, 5*time.Second)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
