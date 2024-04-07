package raylib_adapter

import (
	"gui"
	"gui/color"
	"gui/element"
	"gui/events"
	"gui/fps"
	"gui/utils"
	"gui/window"
	ic "image/color"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func View(data *gui.Window, width, height int32) {
	data.Document.Style["width"] = strconv.Itoa(int(width)) + "px"
	data.Document.Style["height"] = strconv.Itoa(int(height)) + "px"

	wm := window.NewWindowManager()
	wm.FPS = true

	wm.OpenWindow(width, height)
	defer wm.CloseWindow()

	evts := map[string]element.EventList{}

	eventStore := &evts

	// Main game loop
	for !wm.WindowShouldClose() {
		rl.BeginDrawing()

		// Check if the window size has changed
		newWidth := int32(rl.GetScreenWidth())
		newHeight := int32(rl.GetScreenHeight())

		if newWidth != width || newHeight != height {
			rl.ClearBackground(rl.RayWhite)
			// Window has been resized, handle the event
			width = newWidth
			height = newHeight

			data.CSS.Width = float32(width)
			data.CSS.Height = float32(height)

			data.Document.Style["width"] = strconv.Itoa(int(width)) + "px"
			data.Document.Style["height"] = strconv.Itoa(int(height)) + "px"
		}

		eventStore = events.GetEvents(&data.Document.Children[0], eventStore)
		data.CSS.ComputeNodeStyle(&data.Document.Children[0])
		rd := data.CSS.Render(data.Document.Children[0])
		wm.LoadTextures(rd)
		wm.Draw(rd)

		events.RunEvents(eventStore)

		rl.EndDrawing()
	}
}

type Rect struct {
	Node  rl.Rectangle
	Color rl.Color // Added a Color field
	Text  Text
}

type Text struct {
	Color rl.Color
	Size  float32
	Value string
	Font  string
}

// WindowManager manages the window and rectangles
type WindowManager struct {
	Fonts      map[string]rl.Font
	FPS        bool
	FPSCounter fps.FPSCounter
	Textures   map[string]TextTexture
}

type TextTexture struct {
	Text  string
	Image rl.Texture2D
}

// NewWindowManager creates a new WindowManager instance
func NewWindowManager() *WindowManager {
	fpsCounter := fps.NewFPSCounter()

	return &WindowManager{
		Fonts:      make(map[string]rl.Font),
		FPSCounter: *fpsCounter,
	}
}

// OpenWindow opens the window
func (wm *WindowManager) OpenWindow(width, height int32) {
	rl.InitWindow(width, height, "")
	rl.SetTargetFPS(30)
	// Enable window resizing
	rl.SetWindowState(rl.FlagWindowResizable)
}

// CloseWindow closes the window
func (wm *WindowManager) CloseWindow() {
	rl.CloseWindow()
}

func (wm *WindowManager) LoadTextures(nodes []element.Node) {
	if wm.Textures == nil {
		wm.Textures = map[string]TextTexture{}
	}
	for _, node := range nodes {
		if node.Properties.Text.Image != nil {
			if wm.Textures[node.Properties.Id].Text != node.InnerText {
				rl.UnloadTexture(wm.Textures[node.Properties.Id].Image)
				texture := rl.LoadTextureFromImage(rl.NewImageFromImage(node.Properties.Text.Image))
				wm.Textures[node.Properties.Id] = TextTexture{
					Text:  node.InnerText,
					Image: texture,
				}
			}

		}

	}
}

// Draw draws all nodes on the window
func (wm *WindowManager) Draw(nodes []element.Node) {

	for _, node := range nodes {
		bw, _ := utils.ConvertToPixels(node.Properties.Border.Width, node.Properties.EM, node.Properties.Computed["width"])
		rad, _ := utils.ConvertToPixels(node.Properties.Border.Radius, node.Properties.EM, node.Properties.Computed["width"])

		p := utils.GetMP(node, "padding")

		rect := rl.NewRectangle(node.Properties.X+bw,
			node.Properties.Y+bw,
			node.Properties.Computed["width"]-(bw+bw),
			(node.Properties.Computed["height"]+(p.Top+p.Bottom))-(bw+bw),
		)

		rl.DrawRectangleRoundedLines(rect, rad/200, 1000, bw, node.Properties.Border.Color)
		rl.DrawRectangleRounded(rect, rad/200, 1000, color.Parse(node.Style, "background"))

		if node.Properties.Text.Image != nil {
			r, g, b, a := node.Properties.Text.Color.RGBA()
			rl.DrawTexture(wm.Textures[node.Properties.Id].Image, int32(node.Properties.X+p.Left+bw), int32(node.Properties.Y+p.Top), ic.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
		}
	}

	if wm.FPS {
		wm.FPSCounter.Update()
		wm.FPSCounter.Draw(10, 10, 10, rl.DarkGray)
	}

}

// WindowShouldClose returns true if the window should close
func (wm *WindowManager) WindowShouldClose() bool {
	return rl.WindowShouldClose()
}
