package main

import (
	"log"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/cairo"
	"errors"
	"strconv"
)

var filename string
var grey = Color{0.2, 0.2, 0.2}

func openFile() (Puzzle, error) {

	filename = getFilenameFromUser()

	var puzzle Puzzle
	if filename != "" {
		puzzle, err := ReadFile(filename)
		if err != nil {
			return puzzle, err
		}
		return puzzle, nil
	}
	return puzzle, errors.New("User canceled action.")
}

func getFilenameFromUser() string {
	fileChooser, err := gtk.FileChooserDialogNewWith2Buttons(
		"Open Shape Model",
		nil,
		gtk.FILE_CHOOSER_ACTION_OPEN,
		"Cancel", gtk.RESPONSE_DELETE_EVENT,
		"Open", gtk.RESPONSE_ACCEPT)

	if err != nil {
		return ""
	}

	// filter for models
	filter, _ := gtk.FileFilterNew()
	filter.AddPattern("*.model")
	filter.SetName("Shape Models")
	fileChooser.AddFilter(filter)

	switcher := fileChooser.Run()
	filename := fileChooser.GetFilename()
	fileChooser.Destroy()

	// if the user pressed another button other than OK
	if switcher != -3 {
		return ""
	}

	return filename
}

func drawGrid(puzzle Puzzle, grid Grid, cellSize float64, cr *cairo.Context) {

	colors := GenerateColors(len(puzzle.Pieces))

	// draws all the cells
	var i, j int
	for i = 0; i < len(grid); i++ {
		for j = 0; j < len(grid[0]); j++ {
			if grid[i][j] > 0 {
				var number = ""
				if puzzle.WinInfo.DrawNumbers {
					number = strconv.Itoa(int(grid[i][j]))
				}
				drawCell(j, i, cellSize, colors[grid[i][j]-1], cr, number)
			}
		}
	}
}

func drawCell(i int, j int, cellSize float64, color Color, cr *cairo.Context, num string) {

	// computes where to locate the cell
	x := 9 + cellSize*float64(i)
	y := 9 + cellSize*float64(j)

	// draws the cell
	setColor(color, cr)
	cr.Rectangle(x, y, cellSize, cellSize)
	cr.Fill()

	// draws the border of the cell
	setColor(grey, cr)
	DrawRectangle(x, y, cellSize, cellSize, cr, num)
}

func setColor(color Color, context *cairo.Context) {
	context.SetSourceRGB(color.R, color.G, color.B)
}

func CreateAndStartGui(filename string, puzzle Puzzle) {

	puzzle.HasGui = true
	gtk.Init(nil)

	cellSize := float64(150)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Shape Puzzle Solver")
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

	upper := 1000.0
	adj, err := gtk.AdjustmentNew(float64(upper-StartingSpeed), 0.0, upper, 5.0, 0.0, 0.0)
	if err != nil {
		log.Fatal("Unable to create adjustement:", err)
	}
	scale, err := gtk.ScaleNew(gtk.ORIENTATION_HORIZONTAL, adj)
	if err != nil {
		log.Fatal("Unable to create scale:", err)
	}
	scale.SetHExpand(true)
	scale.SetDrawValue(false)

	solveButton, err := gtk.ButtonNewWithLabel("Find new solution")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}

	statusGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create status grid:", err)
	}
	statusGrid.SetOrientation(gtk.ORIENTATION_HORIZONTAL)
	statusGrid.SetBorderWidth(5)

	statusBar, err := gtk.StatusbarNew()
	if err != nil {
		log.Fatal("Unable to create statusbar:", err)
	}
	statusBar.SetBorderWidth(0)
	statusBar.SetMarginBottom(0)
	statusBar.SetMarginTop(0)
	statusBar.SetMarginStart(0)
	statusBar.Push(1, "Ready")

	winInfo := WinInfo{win, *statusBar, *solveButton,  StartingSpeed, false}
	puzzle.WinInfo = &winInfo

	scale.Connect("value-changed", func() {
		winInfo.Speed = uint(adj.GetUpper() - scale.GetValue())
	})

	infoLabel, err := gtk.LabelNew("   Speed")
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}

	solveButton.Connect("clicked", func() {
		if puzzle.IsRunning == false {
			solveButton.SetLabel("Stop Finding")
			puzzle.IsRunning = true
			var solutions []Grid
			puzzle.Solutions = &solutions
			puzzle.Pieces = GetPiecesFromGrid(puzzle.OriginalGrid)
			puzzle.WinInfo.StatusBar.Push(1, "Solving...")
			go Solver(&puzzle)
		} else {
			solveButton.SetLabel("Find solutions")
			puzzle.IsRunning = false
			ReadFile(filename)
		}
	})

	da.SetHExpand(true)
	da.SetVExpand(true)

	da.Connect("draw", func(da *gtk.DrawingArea, cr *cairo.Context) {
		windowWidth := float64(da.GetAllocatedWidth())
		windowHeight := float64(da.GetAllocatedHeight())
		windowRatio := windowWidth / windowHeight

		puzzleWidth := float64(len(puzzle.OriginalGrid[0]))
		puzzleHeight := float64(len(puzzle.OriginalGrid))
		puzzleRatio := puzzleWidth / puzzleHeight

		if windowRatio > puzzleRatio {
			cellSize = (windowHeight-20)/puzzleHeight
		} else {
			cellSize = (windowWidth-20)/puzzleWidth
		}

		// draws background
		cr.SetSourceRGB(1, 1, 1)
		cr.Rectangle(0, 0, windowWidth, windowHeight)
		cr.Fill()

		// draws border
		cr.SetSourceRGB(0.1, 0.1, 0.1)
		DrawRectangle(0, 0, windowWidth, windowHeight, cr, "")

		// draws the gtkGrid
		drawGrid(puzzle, puzzle.WorkingGrid, cellSize, cr)
	})

	// creates menu
	menuBar, err := gtk.MenuBarNew()
	if err != nil {
		log.Fatal("Unable to create menubar:", err)
	}

	fileMenu, err := gtk.MenuNew()
	if err != nil {
		log.Fatal("Unable to create menu:", err)
	}

	fileMenuItem, err := gtk.MenuItemNewWithLabel("File")
	if err != nil {
		log.Fatal("Unable to create menuitem:", err)
	}

	openMenuItem, err := gtk.MenuItemNewWithLabel("Open")
	if err != nil {
		log.Fatal("Unable to create menuitem:", err)
	}

	openMenuItem.Connect("activate", func() {
		newPuzzle, err := openFile()
		if err != nil {
			dialog := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_CLOSE, "An error has occurred opening the model: %s", err.Error())
			dialog.SetModal(true)
			dialog.Connect("response",  func() {
				dialog.Destroy()
			})
			dialog.Show()
			return
		}

		puzzle = newPuzzle
		puzzle.HasGui = true
		puzzle.WinInfo = &winInfo
		solveButton.SetLabel("Find solutions")
		statusBar.Push(1, "Ready")
	})

	fileMenuItem.SetSubmenu(fileMenu)
	fileMenu.Append(openMenuItem)
	menuBar.Append(fileMenuItem)

	viewMenu, err := gtk.MenuNew()
	if err != nil {
		log.Fatal("Unable to create view menu:", err)
	}

	viewMenuItem, err := gtk.MenuItemNewWithLabel("View")
	if err != nil {
		log.Fatal("Unable to create menuitem:", err)
	}

	solutionsMenuItem, err := gtk.MenuItemNewWithLabel("Solutions")
	if err != nil {
		log.Fatal("Unable to create menuitem:", err)
	}

	viewMenuItem.SetSubmenu(viewMenu)
	viewMenu.Append(solutionsMenuItem)
	menuBar.Append(viewMenuItem)

	solutionsMenuItem.Connect("activate", func() {
		ShowSolutions(puzzle)
	})

	gtkGrid.Attach(menuBar, 0, 0, 200, 200)

	gtkGrid.Add(da)
	controlsGrid.Add(solveButton)
	controlsGrid.Add(infoLabel)
	controlsGrid.Add(scale)
	gtkGrid.Add(controlsGrid)

	statusGrid.Add(statusBar)
	gtkGrid.Add(statusGrid)
	win.Add(gtkGrid)
	win.ShowAll()

	gtk.Main()
}
