package chrono

import (
	"context"
	"log"
	"os"
	"os/signal"
	"testing"
	"time"
)

type TestTask1 struct {
}

func (task *TestTask1) Run(ctx context.Context) {
	log.Print("Hello")
	time.Sleep(5 * time.Second)
}

func TestNewScheduledTaskExecutor(t *testing.T) {
	//scheduledTaskExecutor := NewDefaultScheduledExecutor()
	//scheduledTaskExecutor.ScheduleAtWithRate(&TestTask1{}, 0, 8*time.Second)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
