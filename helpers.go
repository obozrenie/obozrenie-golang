package main

import (
	"fmt"
	"time"

	"github.com/skybon/multilogger"
)

func PrettyLogMessage(status int, text string, mtyped multilogger.LogMessageType) multilogger.LogMessage {
	return multilogger.MakeLogMessage(time.Now(), mtyped, fmt.Sprintf("%d: %s", status, text))
}
