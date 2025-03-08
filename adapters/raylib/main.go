package raylib

import (
	adapter "grim/adapters"
	"grim/element"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Init() *adapter.Adapter {
	rl.SetTraceLogLevel(rl.LogNone)
	a := adapter.Adapter{}
	a.AddEventListener("cursor", handleCursorEvent)

	wm := NewWindowManager(&a)
	a.Init = func(width, height int) {
		wm.OpenWindow(int32(width), int32(height))
		a.Library.UnloadCallback = func(key string) {
			t, exists := wm.Textures[key]
			if exists {
				rl.UnloadTexture(*t)
				delete(wm.Textures, key)
			}
		}
	}
	a.Load = wm.LoadTextures
	a.Render = func(state []element.State) {
		if rl.WindowShouldClose() {
			a.DispatchEvent(element.Event{Name: "close"})
		}
		wm.Draw(state)
	}

	fs := adapter.FileSystem{}
	fs.ReadFile = func(path string) ([]byte, error) {
		data, err := os.ReadFile(path)
		return data, err
	}
	fs.WriteFile = func(path string, data []byte) {
		os.WriteFile(path, data, 0644)
	}
	getSystemFonts(&fs)

	a.FileSystem = fs
	return &a
}

// handleCursorEvent sets the mouse cursor based on the event data
func handleCursorEvent(e element.Event) {
	cursorMap := map[string]rl.MouseCursor{
		"":            rl.MouseCursorArrow,
		"text":        rl.MouseCursorIBeam,
		"crosshair":   rl.MouseCursorCrosshair,
		"pointer":     rl.MouseCursorPointingHand,
		"ew-resize":   rl.MouseCursorResizeEW,
		"ns-resize":   rl.MouseCursorResizeNS,
		"nwse-resize": rl.MouseCursorResizeNWSE,
		"nesw-resize": rl.MouseCursorResizeNESW,
		"grab":        9,
		"not-allowed": rl.MouseCursorNotAllowed,
	}

	if cursor, found := cursorMap[e.Data.(string)]; found {
		rl.SetMouseCursor(cursor)
	}
}

// WindowManager manages the window and rectangles
type WindowManager struct {
	FPSCounterOn  bool
	FPS           int32
	Textures      map[string]*rl.Texture2D
	Width         int32
	Height        int32
	CurrentEvents map[int]bool
	MousePosition []int
	MouseState    bool
	ContextState  bool
	Adapter       *adapter.Adapter
}

// NewWindowManager creates a new WindowManager instance
func NewWindowManager(a *adapter.Adapter) *WindowManager {

	mp := rl.GetMousePosition()
	return &WindowManager{
		CurrentEvents: make(map[int]bool, 256),
		MousePosition: []int{int(mp.X), int(mp.Y)},
		Adapter:       a,
	}
}

// OpenWindow opens the window
func (wm *WindowManager) OpenWindow(width, height int32) {
	rl.InitWindow(width, height, "")
	rl.SetTargetFPS(120)
	// rl.SetTargetFPS(60)
	wm.Width = width
	wm.Height = height
	// Enable window resizing
	rl.SetWindowState(rl.FlagWindowResizable)
}

func (wm *WindowManager) LoadTextures(nodes []element.State) {
	if wm.Textures == nil {
		wm.Textures = make(map[string]*rl.Texture2D)
	}

	for _, node := range nodes {
		if len(node.Textures) > 0 {
			for _, key := range node.Textures {
				rt, exists := wm.Textures[key]
				texture, inLibrary := wm.Adapter.Library.Get(key)
				matches := true
				if inLibrary && exists {
					tb := texture.Bounds()
					matches = (rt.Width == int32(tb.Dx()) && rt.Height == int32(tb.Dy()))
				}
				// Unload existing texture if there is a mismatch
				if exists && (!matches || !inLibrary) {
					rl.UnloadTexture(*rt)
					delete(wm.Textures, key)
				}
				if (!exists && inLibrary) || !matches {
					textureLoaded := rl.LoadTextureFromImage(rl.NewImageFromImage(texture))
					// rl.SetTextureFilter(textureLoaded, rl.FilterBilinear)
					wm.Textures[key] = &textureLoaded
				}
			}

		}
	}
}

// Draw draws all nodes on the window
func (wm *WindowManager) Draw(nodes []element.State) {
	indexes := []float32{0}
	rl.BeginDrawing()
	wm.GetEvents()
	for a := 0; a < len(indexes); a++ {
		for _, node := range nodes {
			if node.Hidden {
				continue
			}
			if node.Z == indexes[a] {

				// Draw the border based on the style for each side

				if node.Textures != nil {
					for _, v := range node.Textures {
						texture, exists := wm.Textures[v]
						if exists {
							sourceRec := rl.Rectangle{
								X:      0,
								Y:      0,
								Width:  float32(texture.Width),
								Height: float32(texture.Height),
							}

							if node.Crop.X != 0 || node.Crop.Y != 0 || node.Crop.Width != 0 || node.Crop.Height != 0 {
								sourceRec = rl.Rectangle{
									X:      float32(node.Crop.X),
									Y:      float32(node.Crop.Y),
									Width:  float32(node.Crop.Width),
									Height: float32(node.Crop.Height),
								}
							}

							rl.DrawTextureRec(*texture, sourceRec, rl.Vector2{
								X: node.X + float32(node.Crop.X),
								Y: node.Y + float32(node.Crop.Y),
							}, rl.White)
						}
					}
				}
			} else {
				if !slices.Contains(indexes, node.Z) {
					indexes = append(indexes, node.Z)
					slices.Sort(indexes)
				}
			}
		}
	}

	rl.EndDrawing()
}

func (wm *WindowManager) GetEvents() {
	cw := rl.GetScreenWidth()
	ch := rl.GetScreenHeight()
	if cw != int(wm.Width) || ch != int(wm.Height) {
		e := element.Event{
			Name: "windowresize",
			Data: map[string]int{"width": cw, "height": ch},
		}
		wm.Width = int32(cw)
		wm.Height = int32(ch)
		wm.Adapter.DispatchEvent(e)
	}
	CtrlKey := false
	MetaKey := false
	ShiftKey := false
	AltKey := false

	if rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl) {
		CtrlKey = true
	} else {
		CtrlKey = false
	}

	if rl.IsKeyDown(343) {
		MetaKey = true
	} else {
		MetaKey = false
	}
	if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
		ShiftKey = true
	} else {
		ShiftKey = false
	}
	if rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt) {
		AltKey = true
	} else {
		AltKey = false
	}

	// Other keys
	for i := 0; i <= 350; i++ {
		// for i := 32; i < 126; i++ {
		isDown := rl.IsKeyDown(int32(i))
		if wm.CurrentEvents[i] != isDown {
			if isDown {
				keydown := element.Event{
					Name:     "keydown",
					Data:     i,
					CtrlKey:  CtrlKey,
					MetaKey:  MetaKey,
					ShiftKey: ShiftKey,
					AltKey:   AltKey,
				}

				wm.CurrentEvents[i] = true
				wm.Adapter.DispatchEvent(keydown)
			} else {
				keyup := element.Event{
					Name:     "keyup",
					Data:     i,
					CtrlKey:  CtrlKey,
					MetaKey:  MetaKey,
					ShiftKey: ShiftKey,
					AltKey:   AltKey,
				}
				wm.CurrentEvents[i] = false
				wm.Adapter.DispatchEvent(keyup)
			}
		}
	}
	// mouse move, ctrl, shift etc

	mp := rl.GetMousePosition()
	if wm.MousePosition[0] != int(mp.X) || wm.MousePosition[1] != int(mp.Y) {
		wm.Adapter.DispatchEvent(element.Event{
			Name: "mousemove",
			Data: []int{int(mp.X), int(mp.Y)},
		})
		wm.MousePosition[0] = int(mp.X)
		wm.MousePosition[1] = int(mp.Y)
	}
	md := rl.IsMouseButtonDown(rl.MouseLeftButton)
	if md != wm.MouseState {
		if md {
			wm.Adapter.DispatchEvent(element.Event{
				Name: "mousedown",
			})
			wm.MouseState = true
		} else {
			wm.Adapter.DispatchEvent(element.Event{
				Name: "mouseup",
			})
			wm.MouseState = false
		}
	}

	cs := rl.IsMouseButtonPressed(rl.MouseRightButton)
	if cs != wm.ContextState {
		if cs {
			wm.Adapter.DispatchEvent(element.Event{
				Name: "contextmenudown",
			})
			wm.ContextState = true
		} else {
			wm.Adapter.DispatchEvent(element.Event{
				Name: "contextmenuup",
			})
			wm.ContextState = false
		}
	}

	wd := rl.GetMouseWheelMove()

	if wd != 0 {
		wm.Adapter.DispatchEvent(element.Event{
			Name: "scroll",
			Data: int(wd * 6),
		})
	}
}

func getSystemFonts(fs *adapter.FileSystem) {

	switch runtime.GOOS {
	case "windows":
		// System Fonts
		fs.AddDir("C:\\Windows\\Fonts")
		// User Fonts
		fs.AddDir("%APPDATA%\\Microsoft\\Windows\\Fonts")
	case "darwin":
		// System Fonts
		fs.AddDir("/System/Library/Fonts")
		fs.AddDir("/Library/Fonts")
		// User Fonts
		fs.AddDir(filepath.Join(os.Getenv("HOME"), "Library/Fonts"))
	case "linux":
		// System Fonts
		fs.AddDir("/usr/share/fonts")
		fs.AddDir("/usr/local/share/fonts")
		// User Fonts
		fs.AddDir(filepath.Join(os.Getenv("HOME"), ".fonts"))
	}
}
