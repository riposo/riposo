package riposo

import (
	"log"
	"os"
)

// Logger is the default logger for riposo.
var Logger = log.New(os.Stdout, "", log.LstdFlags)
