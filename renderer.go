package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/altwine/go-mindustry-ping/pkg/serverinfo"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

var globalFontFace *truetype.Font

//go:embed fonts/mindustry.ttf
var fontBytes []byte

func initFont() {
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Printf("ошибка загрузки шрифта: %v", err)
		return
	}
	globalFontFace = font
}

func loadFont(dc *gg.Context, fontSize float64) {
	dc.SetFontFace(truetype.NewFace(globalFontFace, &truetype.Options{Size: fontSize}))
}

const (
	width      = 1200
	height     = 1200
	cardWidth  = 1100
	cardHeight = 1100
)

const (
	mainBgColor    = "#0F0F0F"
	textPrimary    = "#FFFFFF"
	secondaryColor = "#4CAF7A"
	lineColor      = "#66CCFF"
	axisColor      = "#3A3A3A"
	labelColor     = "#AAAAAA"
)

// todo(altwine): refactor this lol
// assumes that current font is already set and size if `fontSize`
func DrawMindustryFormatString(dc *gg.Context, text string, x, y float64) {
	textArr := []rune(text)
	xSum := 0.0
	isBuildingColor := false
	stepByStepColor := ""
	for i := 0; i < len(textArr); i += 1 {
		r := string(textArr[i])
		if r == "[" {
			stepByStepColor = ""
			isBuildingColor = true
			continue
		}
		if r == "]" {
			isBuildingColor = false
			if stepByStepColor == "" {
				dc.SetHexColor(textPrimary)
				continue
			}
			if stepByStepColor[0] == '#' && len(stepByStepColor) <= 7 {
				dc.SetHexColor(stepByStepColor + strings.Repeat("0", 7-len(stepByStepColor)))
				continue
			} else {
				mc, is_valid := serverinfo.MINDUSTRY_COLORS[stepByStepColor]
				if !is_valid {
					log.Printf("lol invalid color: %v (%s), just skippin", mc, stepByStepColor)
					continue
				}
				dc.SetColor(color.RGBA(color.RGBA{
					R: uint8(mc.R),
					G: uint8(mc.G),
					B: uint8(mc.B),
					A: uint8(mc.A),
				}))
				continue
			}
		}
		if isBuildingColor {
			stepByStepColor += r
			continue
		}

		dc.DrawString(r, x+xSum, y)
		w, _ := dc.MeasureString(r)
		xSum += w
	}
}

func measureMindustryString(dc *gg.Context, text string) float64 {
	textArr := []rune(text)
	isBuildingColor := false
	xSum := 0.0
	for i := 0; i < len(textArr); i += 1 {
		r := string(textArr[i])
		if r == "[" {
			isBuildingColor = true
			continue
		}
		if r == "]" {
			isBuildingColor = false
			continue
		}
		if isBuildingColor {
			continue
		}
		w, _ := dc.MeasureString(r)
		xSum += w
	}
	return xSum
}

func genImage(dc *gg.Context, si serverinfo.ServerInfo, hr []HistoryRecord) {
	dc.SetHexColor(mainBgColor)
	dc.Clear()
	cardX := float64(width-cardWidth) / 2
	cardY := float64(height-cardHeight) / 2

	const yAxisWidth = 65.0
	const rightMarginText = 20.0*2 + 23.0

	xMinGraph := cardX + yAxisWidth
	xMaxGraph := cardX + cardWidth

	d := 200.0
	yTop := 120.0 + d
	yBottom := 700.0 + d
	chartHeight := yBottom - yTop

	maxPlayers := 0
	for _, r := range hr {
		if r.Players > maxPlayers {
			maxPlayers = r.Players
		}
	}
	dataMax := float64(((maxPlayers + 4) / 5) * 5)
	if dataMax == 0 {
		dataMax = 5
	}

	mapY := func(players int) float64 {
		val := float64(players)
		if val < 0 {
			val = 0
		}
		if val > dataMax {
			val = dataMax
		}
		ratio := (dataMax - val) / dataMax
		return yTop + ratio*chartHeight
	}

	axisX := xMinGraph - 10
	loadFont(dc, 24)
	for v := 0.0; v <= dataMax; v += 5 {
		y := yTop + (dataMax-v)/dataMax*chartHeight
		label := fmt.Sprintf("%.0f", v)
		w, h := dc.MeasureString(label)

		dc.SetHexColor(axisColor)
		dc.SetLineWidth(1)
		dc.DrawLine(axisX, y, cardX+cardWidth, y)
		dc.Stroke()

		dc.SetHexColor(labelColor)
		dc.DrawString(label, axisX-5-w, y+h/3)
	}

	dc.SetLineWidth(2)
	dc.SetHexColor(lineColor)
	for i := 0; i < len(hr)-1; i++ {
		if hr[i].Players == -1 || hr[i+1].Players == -1 {
			continue
		}
		x1 := xMinGraph + (float64(i)/float64(len(hr)-1))*(xMaxGraph-xMinGraph)
		y1 := mapY(hr[i].Players)
		x2 := xMinGraph + (float64(i+1)/float64(len(hr)-1))*(xMaxGraph-xMinGraph)
		y2 := mapY(hr[i+1].Players)
		dc.DrawLine(x1, y1, x2, y2)
		dc.Stroke()
	}

	const maxTimeLabels = 12
	if len(hr) > 0 {
		loadFont(dc, 24)
		dc.SetHexColor(labelColor)
		for i := 0; i < maxTimeLabels; i++ {
			t := float64(i) / float64(maxTimeLabels-1)
			idx := int(math.Round(t * float64(len(hr)-1)))
			record := hr[idx]
			timeStr := time.Unix(record.Timestamp, 0).Format("15:04")
			w, _ := dc.MeasureString(timeStr)
			x := xMinGraph + t*(xMaxGraph-xMinGraph-rightMarginText)
			dc.DrawString(timeStr, x-w/2, yBottom+40)
		}
	}

	dc.SetHexColor(textPrimary)
	currSize := 128.0
	yDec := 0.0
	for {
		loadFont(dc, currSize)
		strWidth := measureMindustryString(dc, si.Host)
		if strWidth+cardX > cardWidth {
			currSize -= 8.0
			yDec += 4.0
		} else {
			break
		}
	}
	DrawMindustryFormatString(dc, si.Host, cardX, cardY+105-yDec)

	loadFont(dc, 58)
	dc.SetHexColor(labelColor)
	dc.DrawString(si.Address, cardX, cardY+185)

	// online
	var sumPlayers int
	var countValid int
	for _, r := range hr {
		if r.Players != -1 {
			sumPlayers += r.Players
			countValid++
		}
	}
	var avgStr string
	if countValid == 0 {
		avgStr = "0"
	} else {
		averagePlayers := float64(sumPlayers) / float64(countValid)
		avgStr = fmt.Sprintf("%.1f", averagePlayers)
	}

	// ping
	var sumPing int
	var countValid2 int
	for _, r := range hr {
		if r.Ping != -1 {
			sumPing += r.Ping
			countValid2++
		}
	}
	var avgStr2 string
	if countValid2 == 0 {
		avgStr2 = "0"
	} else {
		averagePing := float64(sumPing) / float64(countValid2)
		avgStr2 = fmt.Sprintf("%.0f", averagePing)
	}

	yCommon := 65.0

	loadFont(dc, 128)
	dc.SetHexColor(secondaryColor)
	dc.DrawString(avgStr, cardX+40, yCommon+yBottom+140+20)

	loadFont(dc, 32)
	dc.SetHexColor(textPrimary)
	dc.DrawString("средний онлайн:", cardX+40, yCommon+yBottom+50)

	loadFont(dc, 128)
	dc.SetHexColor(secondaryColor)
	str2 := strconv.Itoa(maxPlayers)
	wStr2, _ := dc.MeasureString(str2)
	dc.DrawString(str2, cardX+cardWidth/2-wStr2, yCommon+yBottom+140+20)

	loadFont(dc, 32)
	dc.SetHexColor(textPrimary)
	dc.DrawString("макс онлайн:", cardX+cardWidth/2-wStr2, yCommon+yBottom+50)

	loadFont(dc, 128)
	dc.SetHexColor(secondaryColor)
	str3 := avgStr2
	wStr3, _ := dc.MeasureString(str3)
	dc.DrawString(str3, cardWidth-cardX-wStr3/2-40, yCommon+yBottom+140+20)

	loadFont(dc, 32)
	dc.SetHexColor(textPrimary)
	dc.DrawString("средний пинг:", cardWidth-cardX-wStr3/2-40, yCommon+yBottom+50)
}
