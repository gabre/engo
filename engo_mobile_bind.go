//+build mobilebind

package engo

import (
	"errors"
	"io"
	"runtime"
	"time"

	"engo.io/gl"
	mgl "golang.org/x/mobile/gl"
)

var (
	// Gl is the current OpenGL context
	Gl     *gl.Context
	worker mgl.Worker

	canvasWidth, canvasHeight float32

	msaaPreference int

	ResizeXOffset = float32(0)
	ResizeYOffset = float32(0)

	Backend string = "Mobile"

	drawEvent  = make(chan struct{})
	drawDone   = make(chan struct{})
	initalized = false
)

// CreateWindow creates a window with the specified parameters
func CreateWindow(title string, width, height int, fullscreen bool, msaa int) {
	gameWidth = float32(width)
	gameHeight = float32(height)
	msaaPreference = msaa
}

// WindowSize returns the width and height of the current window
func WindowSize() (w, h int) {
	return int(windowWidth), int(windowHeight)
}

// CursorPos returns the current cursor position
func CursorPos() (x, y float32) {
	notImplemented("CursorPos")
	return 0, 0
}

// WindowWidth returns the current window width
func WindowWidth() float32 {
	return windowWidth
}

// WindowHeight returns the current window height
func WindowHeight() float32 {
	return windowHeight
}

// CanvasWidth returns the current canvas width
func CanvasWidth() float32 {
	return canvasWidth
}

// CanvasHeight returns the current canvas height
func CanvasHeight() float32 {
	return canvasHeight
}

func CanvasScale() float32 {
	return CanvasWidth() / WindowWidth()
}

func DestroyWindow() { /* nothing to do here? */ }

func runLoop(defaultScene Scene, headless bool) {
	go func() {
		for {
			mobileDraw(defaultScene)
		}
	}()
}

// RunPreparation is called only once, and is called automatically when calling Open
// It is only here for benchmarking in combination with OpenHeadlessNoRun
func RunPreparation(defaultScene Scene) {
	windowWidth = float32(opts.MobileWidth)
	canvasWidth = float32(opts.MobileWidth)
	windowHeight = float32(opts.MobileHeight)
	canvasHeight = float32(opts.MobileHeight)
	ResizeXOffset = gameWidth - canvasWidth
	ResizeYOffset = gameHeight - canvasHeight

	Gl.Viewport(0, 0, opts.MobileWidth, opts.MobileHeight)

	Time = NewClock()
	SetScene(defaultScene, false)
}

// RunIteration runs every time android calls to update the screen
func RunIteration() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	select {
	case drawEvent <- struct{}{}:
		for {
			select {
			case <-worker.WorkAvailable():
				worker.DoWork()
			case <-drawDone:
				Input.Mouse.Action = Neutral
				return
			}
		}
	case <-time.After(500 * time.Millisecond):
		Input.Mouse.Action = Neutral
		return
	}
}

// SetCursor changes the cursor - not yet implemented
func SetCursor(Cursor) {
	notImplemented("SetCursor")
}

//SetCursorVisibility sets the visibility of the cursor.
//If true the cursor is visible, if false the cursor is not.
//Does nothing in mobile since there's no visible cursor to begin with
func SetCursorVisibility(visible bool) {}

// SetTitle has no effect on mobile
func SetTitle(title string) {}

// openFile is the mobile-specific way of opening a file
func openFile(url string) (io.ReadCloser, error) {
	return nil, errors.New("binding does not open files this way. utilize go-bindata instead")
}

// mobileDraw runs once per frame. RunIteration for the other backends
func mobileDraw(defaultScene Scene) {
	if !initalized {
		var ctx mgl.Context
		ctx, worker = mgl.NewContext()
		Gl = gl.NewContext(ctx)
	}
	<-drawEvent
	defer func() {
		drawDone <- struct{}{}
	}()

	if !initalized {
		RunPreparation(defaultScene)
		initalized = true
	}

	Time.Tick()

	if !opts.HeadlessMode {
		Input.update()
	}

	// Then update the world and all Systems
	currentWorld.Update(Time.Delta())
}

//TouchEvent handles the touch events sent from Android and puts them in the InputManager
func TouchEvent(x, y, action int) {
	Input.Mouse.X = float32(x)
	Input.Mouse.Y = float32(y)
	switch action {
	case 0:
		Input.Mouse.Action = Press
	case 1:
		Input.Mouse.Action = Release
	case 2:
		Input.Mouse.Action = Move
	}
}

//MobileStop handles when the game is closed
func MobileStop() {
	closeEvent()
	Gl = nil
	worker = nil
}
