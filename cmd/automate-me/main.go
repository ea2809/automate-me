package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ea2809/automate-me/internal/app"
	"github.com/ea2809/automate-me/internal/ui"
)

func main() {
	if err := app.Run(os.Args, ui.NewBubbleUI()); err != nil {
		if errors.Is(err, app.ErrUserCanceled) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
