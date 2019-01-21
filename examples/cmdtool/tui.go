package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/jroimartin/gocui"
)

var (
	clientGui *gocui.Gui
)

// Configure and run the text-based UI
func runTUI() error {

	// Set function to manage all views and keybindings
	clientGui.SetManagerFunc(layout)

	// Bind keys with functions
	clientGui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	clientGui.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, send)

	displayText(strings.Trim(fmt.Sprintf("%sUse '\\' to get all available commands listed\n", prompt), "\n"))
	displayText(prompt)

	// Start main event loop of the TUI
	return clientGui.MainLoop()
}

// Content to be displayed in the TUI
func layout(g *gocui.Gui) error {

	// Get rid of warnings
	_ = g

	// Get size of the terminal
	maxX, maxY := clientGui.Size()

	// Creates view "messages"
	messages, err := clientGui.SetView("messages", 0, 0, maxX-1, maxY-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		messages.Autoscroll = true
		messages.Wrap = true
	}

	// Creates view "input"
	input, err := clientGui.SetView("input", 0, maxY-4, maxX-1, maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		input.Wrap = true
		input.Editable = true
	}

	// Set view "input" as the current view with focus and cursor
	_, err = clientGui.SetCurrentView("input")
	if err != nil {
		return err
	}

	// Show cursor
	clientGui.Cursor = true

	return nil
}

// Quit the TUI
func quit(g *gocui.Gui, v *gocui.View) error {

	// Get rid of warnings
	_, _ = g, v

	//// Unsubscribe via PublisherClient
	//Unsubscribe()

	// Last entry in the individual log file
	//log.Println("Session closing")

	// Last entry in the common log file
	//cLog.Printf("Session closing - details in %q\n", logfilename)

	return gocui.ErrQuit
}

// Send content from the view "input"
func send(g *gocui.Gui, inputView *gocui.View) error {

	// Get rid of warnings
	_ = g

	input := strings.Trim(inputView.Buffer(), "\n")

	// Distinguish between command and chat mode by '\'-prefix
	if strings.HasPrefix(input, "\\") {

		// Interpret "input" as command
		executeCommand(input)

	}
	//else {
	//	// Send "input" to publish
	//	Publish(input)
	//}

	// Clear the "input" and reset the cursor
	inputView.Clear()
	if err := inputView.SetCursor(0, 0); err != nil {
		log.Fatal(err)
	}
	return nil
}

// Display text in "messages"
func displayText(txt string) error {

	// Update the "messages" view as soon as possible
	clientGui.Update(func(g *gocui.Gui) error {
		messagesView, err := clientGui.View("messages")
		if err != nil {
			return fmt.Errorf("could not display text: %v\n", err)
		}
		fmt.Fprintln(messagesView, txt)
		return nil
	})
	return nil
}
