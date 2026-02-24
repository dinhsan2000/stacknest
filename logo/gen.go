//go:build ignore

// gen.go renders the Stacknest logo and writes:
//   build/appicon.png        — 512×512 PNG (Wails app icon)
//   build/windows/icon.ico   — multi-size ICO (16/32/48/256 px)
//
// Run with:  go run logo/gen.go

package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// ─── Bezier / drawing helpers ─────────────────────────────────────────────────

func bezierPt(p0, ctrl, p2 [2]float64, t float64) [2]float64 {
	mt := 1 - t
	return [2]float64{
		mt*mt*p0[0] + 2*mt*t*ctrl[0] + t*t*p2[0],
		mt*mt*p0[1] + 2*mt*t*ctrl[1] + t*t*p2[1],
	}
}

func lerpC(a, b color.NRGBA, t float64) color.NRGBA {
	return color.NRGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: 255,
	}
}

// drawArc renders a thick quadratic bezier stroke with a left→right gradient.
func drawArc(img *image.NRGBA, p0, ctrl, p2 [2]float64, radius float64, colL, colR color.NRGBA) {
	const steps = 600
	b := img.Bounds()
	sw := float64(b.Max.X)

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		pt := bezierPt(p0, ctrl, p2, t)
		c := lerpC(colL, colR, pt[0]/sw)

		xMin := int(pt[0] - radius - 1.5)
		xMax := int(pt[0] + radius + 1.5)
		yMin := int(pt[1] - radius - 1.5)
		yMax := int(pt[1] + radius + 1.5)

		for y := yMin; y <= yMax; y++ {
			if y < b.Min.Y || y >= b.Max.Y {
				continue
			}
			for x := xMin; x <= xMax; x++ {
				if x < b.Min.X || x >= b.Max.X {
					continue
				}
				dx := float64(x) + 0.5 - pt[0]
				dy := float64(y) + 0.5 - pt[1]
				dist := math.Sqrt(dx*dx + dy*dy)

				var alpha float64
				if dist < radius-0.5 {
					alpha = 1.0
				} else if dist < radius+0.5 {
					alpha = radius + 0.5 - dist
				} else {
					continue
				}

				fa := alpha * float64(c.A) / 255.0
				ex := img.NRGBAAt(x, y)
				img.SetNRGBA(x, y, color.NRGBA{
					R: uint8(float64(ex.R)*(1-fa) + float64(c.R)*fa),
					G: uint8(float64(ex.G)*(1-fa) + float64(c.G)*fa),
					B: uint8(float64(ex.B)*(1-fa) + float64(c.B)*fa),
					A: ex.A,
				})
			}
		}
	}
}

// applyRoundedCorners zeros pixels outside the rounded rect.
func applyRoundedCorners(img *image.NRGBA, cornerR float64) {
	sz := float64(img.Bounds().Max.X)
	for y := 0; y < img.Bounds().Max.Y; y++ {
		for x := 0; x < img.Bounds().Max.X; x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			var outside bool
			switch {
			case fx < cornerR && fy < cornerR:
				dx, dy := cornerR-fx, cornerR-fy
				outside = math.Sqrt(dx*dx+dy*dy) > cornerR
			case fx > sz-cornerR && fy < cornerR:
				dx, dy := fx-(sz-cornerR), cornerR-fy
				outside = math.Sqrt(dx*dx+dy*dy) > cornerR
			case fx < cornerR && fy > sz-cornerR:
				dx, dy := cornerR-fx, fy-(sz-cornerR)
				outside = math.Sqrt(dx*dx+dy*dy) > cornerR
			case fx > sz-cornerR && fy > sz-cornerR:
				dx, dy := fx-(sz-cornerR), fy-(sz-cornerR)
				outside = math.Sqrt(dx*dx+dy*dy) > cornerR
			}
			if outside {
				img.SetNRGBA(x, y, color.NRGBA{})
			}
		}
	}
}

// ─── Logo renderer ────────────────────────────────────────────────────────────

// renderLogo returns a Stacknest icon at the given pixel size.
// The design is three nested arcs (widest at bottom) on a dark background.
func renderLogo(size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	s := float64(size)
	k := s / 512.0 // scale factor (arcs defined on 512×512 canvas)

	// Fill background — dark blue (#0e1220)
	bg := color.NRGBA{R: 14, G: 18, B: 32, A: 255}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.SetNRGBA(x, y, bg)
		}
	}

	// Rounded corners (96px radius at 512px → 18.75%)
	applyRoundedCorners(img, s*0.1875)

	// Stroke half-width (44px stroke at 512px)
	r := 22.0 * k

	// ── Three arcs, coordinates from logo.svg ────────────────────────────────
	//                   p0                 ctrl              p2
	// Bottom arc  M68 415    Q256 265    444 415
	// Middle arc  M118 338   Q256 202    394 338
	// Top arc     M168 262   Q256 142    344 262

	// Bottom arc — deepest blue/violet
	drawArc(img,
		[2]float64{68 * k, 415 * k}, [2]float64{256 * k, 265 * k}, [2]float64{444 * k, 415 * k},
		r,
		color.NRGBA{R: 29, G: 78, B: 216, A: 255},  // #1d4ed8 blue-700
		color.NRGBA{R: 109, G: 40, B: 217, A: 255},  // #6d28d9 violet-700
	)
	// Middle arc
	drawArc(img,
		[2]float64{118 * k, 338 * k}, [2]float64{256 * k, 202 * k}, [2]float64{394 * k, 338 * k},
		r,
		color.NRGBA{R: 59, G: 130, B: 246, A: 255},  // #3b82f6 blue-500
		color.NRGBA{R: 139, G: 92, B: 246, A: 255},  // #8b5cf6 violet-500
	)
	// Top arc — lightest
	drawArc(img,
		[2]float64{168 * k, 262 * k}, [2]float64{256 * k, 142 * k}, [2]float64{344 * k, 262 * k},
		r,
		color.NRGBA{R: 147, G: 197, B: 253, A: 255},  // #93c5fd blue-300
		color.NRGBA{R: 196, G: 181, B: 253, A: 255},  // #c4b5fd violet-300
	)

	return img
}

// ─── ICO writer ───────────────────────────────────────────────────────────────

// writeICO packages PNG-compressed images into a Windows ICO file.
// Windows Vista+ supports PNG-in-ICO (RFC 2397 style).
func writeICO(path string, images []*image.NRGBA) error {
	type entry struct {
		w, h     byte
		imgBytes []byte
	}

	var entries []entry
	for _, img := range images {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return err
		}
		sz := img.Bounds().Max.X
		var w, h byte
		if sz >= 256 {
			w, h = 0, 0 // 0 encodes as 256 in the ICO spec
		} else {
			w, h = byte(sz), byte(sz)
		}
		entries = append(entries, entry{w: w, h: h, imgBytes: buf.Bytes()})
	}

	var out bytes.Buffer
	le := binary.LittleEndian
	bw := func(v any) { binary.Write(&out, le, v) } //nolint:errcheck

	// ICONDIR header
	bw(uint16(0))              // reserved
	bw(uint16(1))              // type: 1 = ICO
	bw(uint16(len(entries)))

	// Offset to first image data
	offset := uint32(6 + len(entries)*16)

	// ICONDIRENTRY × N
	for _, e := range entries {
		out.WriteByte(e.w) // width  (0 = 256)
		out.WriteByte(e.h) // height (0 = 256)
		out.WriteByte(0)   // color count (0 = truecolor)
		out.WriteByte(0)   // reserved
		bw(uint16(1))      // planes
		bw(uint16(32))     // bits per pixel
		bw(uint32(len(e.imgBytes)))
		bw(offset)
		offset += uint32(len(e.imgBytes))
	}

	// Image data
	for _, e := range entries {
		out.Write(e.imgBytes)
	}

	return os.WriteFile(path, out.Bytes(), 0644)
}

// ─── main ─────────────────────────────────────────────────────────────────────

func main() {
	// build/appicon.png — 512×512
	logo512 := renderLogo(512)
	f, err := os.Create("build/appicon.png")
	if err != nil {
		panic(err)
	}
	if err := png.Encode(f, logo512); err != nil {
		panic(err)
	}
	f.Close()

	// build/windows/icon.ico — 16, 32, 48, 256 px
	sizes := []int{256, 48, 32, 16}
	var imgs []*image.NRGBA
	for _, sz := range sizes {
		imgs = append(imgs, renderLogo(sz))
	}
	if err := writeICO("build/windows/icon.ico", imgs); err != nil {
		panic(err)
	}
}
