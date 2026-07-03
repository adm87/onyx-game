package tiled

import (
	"encoding/xml"
)

type Orientation string

const (
	OrientationOrthogonal Orientation = "orthogonal"
	OrientationIsometric  Orientation = "isometric"
	OrientationStaggered  Orientation = "staggered"
	OrientationHexagonal  Orientation = "hexagonal"
)

type RenderOrder string

const (
	RenderOrderRightDown RenderOrder = "right-down"
	RenderOrderRightUp   RenderOrder = "right-up"
	RenderOrderLeftDown  RenderOrder = "left-down"
	RenderOrderLeftUp    RenderOrder = "left-up"
)

type Encoding string

const (
	EncodingCSV    Encoding = "csv"
	EncodingBase64 Encoding = "base64"
)

type Compression string

const (
	CompressionGzip Compression = "gzip"
	CompressionZlib Compression = "zlib"
	CompressionZstd Compression = "zstd"
	CompressionNone Compression = ""
)

type TmxObjectGroups []TmxObjectGroup

func (g TmxObjectGroups) EachInGroup(groupName string, fn func(object *TmxObject)) {
	for i := range g {
		group := &g[i]
		if group.Name == groupName {
			for j := range group.Objects {
				object := &group.Objects[j]
				fn(object)
			}
			return
		}
	}
}

type Tmx struct {
	Handle       uint64          `xml:"-"`
	Version      string          `xml:"version,attr"`
	TiledVersion string          `xml:"tiledversion,attr"`
	Orientation  Orientation     `xml:"orientation,attr"`
	RenderOrder  RenderOrder     `xml:"renderorder,attr"`
	Infinite     bool            `xml:"infinite,attr"`
	Width        int             `xml:"width,attr"`
	Height       int             `xml:"height,attr"`
	TileWidth    int             `xml:"tilewidth,attr"`
	TileHeight   int             `xml:"tileheight,attr"`
	NextLayerID  int             `xml:"nextlayerid,attr"`
	NextObjectID int             `xml:"nextobjectid,attr"`
	Tilesets     []TmxTileset    `xml:"tileset"`
	Layers       []TmxLayer      `xml:"layer"`
	ObjectGroups TmxObjectGroups `xml:"objectgroup"`
}

type TmxTileset struct {
	Handle   uint64 `xml:"-"`
	FirstGID int    `xml:"firstgid,attr"`
	Source   string `xml:"source,attr"`
}

type TmxLayer struct {
	ID      int          `xml:"id,attr"`
	Name    string       `xml:"name,attr"`
	Width   int          `xml:"width,attr"`
	Height  int          `xml:"height,attr"`
	Visible bool         `xml:"visible,attr"`
	Data    TmxLayerData `xml:"data"`
}

type TmxLayerData struct {
	Encoding    Encoding        `xml:"encoding,attr"`
	Compression Compression     `xml:"compression,attr"`
	Content     string          `xml:",chardata"`
	Chunks      []TmxLayerChunk `xml:"chunk"`
}

type TmxLayerChunk struct {
	X       int    `xml:"x,attr"`
	Y       int    `xml:"y,attr"`
	Width   int    `xml:"width,attr"`
	Height  int    `xml:"height,attr"`
	Content string `xml:",chardata"`
}

type TmxObjectGroup struct {
	ID      int         `xml:"id,attr"`
	Name    string      `xml:"name,attr"`
	Objects []TmxObject `xml:"object"`
}

type TmxObject struct {
	ID     int     `xml:"id,attr"`
	Name   string  `xml:"name,attr"`
	X      float64 `xml:"x,attr"`
	Y      float64 `xml:"y,attr"`
	Width  float64 `xml:"width,attr"`
	Height float64 `xml:"height,attr"`
}

type Tsx struct {
	Handle       uint64    `xml:"-"`
	Version      string    `xml:"version,attr"`
	TiledVersion string    `xml:"tiledversion,attr"`
	Name         string    `xml:"name,attr"`
	TileWidth    int       `xml:"tilewidth,attr"`
	TileHeight   int       `xml:"tileheight,attr"`
	TileCount    int       `xml:"tilecount,attr"`
	Columns      int       `xml:"columns,attr"`
	Image        TsxImage  `xml:"image"`
	TileOffset   TsxOffset `xml:"tileoffset"`
}

type TsxImage struct {
	Handle uint64 `xml:"-"`
	Source string `xml:"source,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
}

type TsxOffset struct {
	X int `xml:"x,attr"`
	Y int `xml:"y,attr"`
}

func (t *Tmx) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type tmxAlias Tmx

	if err := d.DecodeElement((*tmxAlias)(t), &start); err != nil {
		return err
	}

	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "infinite":
			t.Infinite = attr.Value == "1"
		}
	}

	return nil
}

func (l *TmxLayer) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type layerAlias TmxLayer

	if err := d.DecodeElement((*layerAlias)(l), &start); err != nil {
		return err
	}

	l.Visible = true // Default to visible if the attribute is not present

	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "visible":
			l.Visible = attr.Value == "1"
		}
	}

	return nil
}

func NearestTileset(sets []TmxTileset, gid uint32) (*TmxTileset, int) {
	var nearest *TmxTileset
	var index int
	for i := range sets {
		if sets[i].FirstGID <= int(gid) {
			nearest = &sets[i]
			index = i
		} else {
			break
		}
	}
	return nearest, index
}
