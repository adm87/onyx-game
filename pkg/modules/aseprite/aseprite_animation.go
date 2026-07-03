package aseprite

import (
	"encoding/json"
	"fmt"
	"image/color"

	"github.com/adm87/onyx/pkg/engine/file"
)

type Direction string

const (
	DirectionForward  Direction = "forward"
	DirectionReverse  Direction = "reverse"
	DirectionPingPong Direction = "pingpong"
)

type BlendMode string

const (
	BlendModeNormal   BlendMode = "normal"
	BlendModeAdd      BlendMode = "add"
	BlendModeMultiply BlendMode = "multiply"
	BlendModeScreen   BlendMode = "screen"
)

type AnimationData struct {
	Meta   Meta        `json:"meta"`
	Frames []FrameData `json:"frames"`
}

type Meta struct {
	Clips map[string]FrameTag `json:"_"`

	App       string        `json:"app"`
	Version   string        `json:"version"`
	Format    string        `json:"format"`
	Image     file.FilePath `json:"image"`
	Scale     string        `json:"scale"`
	Size      Frame         `json:"size"`
	FrameTags []FrameTag    `json:"frameTags"`
	Layers    []Layer       `json:"layers"`
	Slices    []any         `json:"slices"` // TODO - add this when needed
}

type FrameTag struct {
	Name      string     `json:"name"`
	From      int        `json:"from"`
	To        int        `json:"to"`
	Direction Direction  `json:"direction"`
	Color     color.RGBA `json:"color"`
}

type Layer struct {
	Name      string    `json:"name"`
	Opacity   int       `json:"opacity"`
	BlendMode BlendMode `json:"blendMode"`
}

type FrameData struct {
	Duration         int    `json:"duration"`
	Filename         string `json:"filename"`
	Rotated          bool   `json:"rotated"`
	Trimmed          bool   `json:"trimmed"`
	SpriteSourceSize Frame  `json:"spriteSourceSize"`
	SourceSize       Frame  `json:"sourceSize"`
	Frame            Frame  `json:"frame"`
}

type Frame struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

func (ft *FrameTag) UnmarshalJSON(data []byte) error {
	type Alias FrameTag
	aux := &struct {
		Color string `json:"color"`
		*Alias
	}{
		Alias: (*Alias)(ft),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	c, err := parseHexColor(aux.Color)
	if err != nil {
		return err
	}
	ft.Color = c

	return nil
}

func parseHexColor(s string) (color.RGBA, error) {
	var c color.RGBA
	_, err := fmt.Sscanf(s, "#%02x%02x%02x%02x", &c.R, &c.G, &c.B, &c.A)
	if err != nil {
		return c, fmt.Errorf("failed to parse hex color '%s': %v", s, err)
	}
	return c, nil
}
