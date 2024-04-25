package helper

import (
	"errors"
	"fmt"
	"time"

	"github.com/sx-network/sx-reporter/command"
	"github.com/sx-network/sx-reporter/infra/common"
)

type ClientCloseResult struct {
	Message string `json:"message"`
}

func (r *ClientCloseResult) GetOutput() string {
	return r.Message
}

// HandleSignals is a helper method for handling signals sent to the console
// Like stop, error, etc.
func HandleSignals(
	closeFn func(),
	outputter command.OutputFormatter,
) error {
	signalCh := common.GetTerminationSignalCh()
	sig := <-signalCh

	closeMessage := fmt.Sprintf("\n[SIGNAL] Caught signal: %v\n", sig)
	closeMessage += "Gracefully shutting down client...\n"

	outputter.SetCommandResult(
		&ClientCloseResult{
			Message: closeMessage,
		},
	)
	outputter.WriteOutput()

	// Call the Minimal server close callback
	gracefulCh := make(chan struct{})

	go func() {
		if closeFn != nil {
			closeFn()
		}

		close(gracefulCh)
	}()

	select {
	case <-signalCh:
		return errors.New("shutdown by signal channel")
	case <-time.After(5 * time.Second):
		return errors.New("shutdown by timeout")
	case <-gracefulCh:
		return nil
	}
}
