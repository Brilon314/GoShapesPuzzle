package main

import (
	"log"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/cairo"
	"fmt"
)

var index int

func ShowSolutions(puzzle Puzzle) {

	gtk.Init(nil)

	cellSize := float64(150)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Solutions")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	//win.SetPosition(gtk.WIN_POS_CENTER)
	width, height := 340, 400
	win.SetDefaultSize(width, height)

	// Create a new gtkGrid widget to arrange child widgets
	gtkGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create gtkGrid:", err)
	}
	gtkGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	gtkGrid.SetBorderWidth(5)

	// Create some widgets to put in the gtkGrid.
	da, err := gtk.DrawingAreaNew()
	if err != nil {
		log.Fatal("Unable to create drawingarea:", err)
	}

	controlsGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create statusbar:", err)
	}
	controlsGrid.SetOrientation(gtk.ORIENTATION_HORIZONTAL)
	controlsGrid.SetBorderWidth(5)

	label, err := gtk.LabelNew(getMessage(puzzle))
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}
	label.SetMarginStart(10)
	index = 0
	btnLeft, err := gtk.ButtonNewWithLabel("<")
	if err != nil {
		log.Fatal("Unable to create button left:", err)
	}
	btnLeft.SetSensitive(false)

	btnRight, err := gtk.ButtonNewWithLabel(">")
	if err != nil {
		log.Fatal("Unable to create button right:", err)
	}
	btnRight.SetSensitive(index < len(*puzzle.Solutions)-1)

	btnLeft.Connect("clicked", func() {
		if index > 0 {
			index --
		}
		btnLeft.SetSensitive(index > 0)
		btnRight.SetSensitive(index < len(*puzzle.Solutions)-1)
		label.SetText(getMessage(puzzle))
		win.QueueDraw()
	})

	btnRight.Connect("clicked", func() {
		if index < len(*puzzle.Solutions)-1 {
			index ++
		}
		btnRight.SetSensitive(index < len(*puzzle.Solutions)-1)
		btnLeft.SetSensitive(index > 0)
		label.SetText(getMessage(puzzle))
		win.QueueDraw()
	})

	da.SetHExpand(true)
	da.SetVExpand(true)

	da.Connect("draw", func(da *gtk.DrawingArea, cr *cairo.Context) {
		windowWidth := float64(da.GetAllocatedWidth())
		windowHeight := float64(da.GetAllocatedHeight())
		windowRatio := windowWidth / windowHeight

		puzzleWidth := float64(len(puzzle.Grid[0]))
		puzzleHeight := float64(len(puzzle.Grid))
		puzzleRatio := puzzleWidth / puzzleHeight

		if windowRatio > puzzleRatio {
			cellSize = (windowHeight - 20) / puzzleHeight
		} else {
			cellSize = (windowWidth - 20) / puzzleWidth
		}

		// draws background
		cr.SetSourceRGB(1, 1, 1)
		cr.Rectangle(0, 0, windowWidth, windowHeight)
		cr.Fill()

		// draws border
		cr.SetSourceRGB(0.1, 0.1, 0.1)
		DrawRectangle(0, 0, windowWidth, windowHeight, cr, "")

		// draws the gtkGrid
		if len(*puzzle.Solutions) > 0 {
			drawGrid(puzzle, (*puzzle.Solutions)[index], cellSize, cr)
		} else {
			label.SetText("No solutions found yet")
		}
	})

	gtkGrid.Add(da)
	controlsGrid.Add(btnLeft)
	controlsGrid.Add(btnRight)
	controlsGrid.Add(label)
	gtkGrid.Add(controlsGrid)
	win.Add(gtkGrid)
	win.ShowAll()

	gtk.Main()
}

func getMessage(puzzle Puzzle) string {
	return fmt.Sprintf("Solution #%d / %d", index+1, len(*puzzle.Solutions))
}
