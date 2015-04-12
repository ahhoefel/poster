package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

type Heatmap interface {
	Size() Size
	At(x, y uint) uint
	Set(x, y, val uint)
}

type Size struct {
	Width, Height uint
}

type HeatmapArray struct {
	slice  []uint
	spread uint
	size   Size
}

func (h HeatmapArray) At(x, y uint) uint {
	return h.slice[x+y*h.spread]
}

func (h HeatmapArray) Set(x, y, val uint) {
	h.slice[x+y*h.spread] = val
}

func (h HeatmapArray) Size() Size {
	return h.size
}

type action func(value uint, neighbors [4]uint) uint

type HeatmapBuffer interface {
	Heatmap() Heatmap
	Buffer() Heatmap
	Swap()
	Apply(a action, boundary uint)
}

type HeatmapBufferArray struct {
	h *HeatmapArray
	b *HeatmapArray
}

func (h HeatmapBufferArray) Heatmap() Heatmap {
	return h.h
}

func (h HeatmapBufferArray) Buffer() Heatmap {
	return h.b
}

func (h *HeatmapBufferArray) Swap() {
	tmp := h.h
	h.h = h.b
	h.b = tmp
}

func (hba *HeatmapBufferArray) Apply(a action, boundary uint) {
	h := hba.h
	for j := range h.slice {
		var neighbours [4]uint
		i := uint(j)
		if i%h.spread == 0 {
			neighbours[0] = boundary
		} else {
			neighbours[0] = h.slice[i-1]
		}
		if i%h.spread == h.spread-1 {
			neighbours[1] = boundary
		} else {
			neighbours[1] = h.slice[i+1]
		}
		if int(i-h.spread) < 0 {
			neighbours[2] = boundary
		} else {
			neighbours[2] = h.slice[i-h.spread]
		}
		if i+h.spread >= uint(len(h.slice)) {
			neighbours[3] = boundary
		} else {
			neighbours[3] = h.slice[i+h.spread]
		}
		hba.b.slice[i] = a(h.slice[i], neighbours)
	}
	hba.Swap()
}

func spreadAction(value uint, neighbors [4]uint) uint {
	for _, v := range neighbors {
		if v != 0 {
			fmt.Print("!")
			return v
		}
	}
	return value
}

func NewHeatmapArray(width, height uint) *HeatmapArray {
	h := new(HeatmapArray)
	*h = HeatmapArray{
		slice:  make([]uint, width*height),
		spread: height,
		size: Size{
			Width:  width,
			Height: height,
		},
	}
	return h
}

func NewHeatmapBufferArray(width, height uint) HeatmapBuffer {
	hba := new(HeatmapBufferArray)
	*hba = HeatmapBufferArray{
		h: NewHeatmapArray(width, height),
		b: NewHeatmapArray(width, height),
	}
	return hba
}

func MakeImage(heatmap Heatmap) image.Image {
	fmt.Printf("Making image %d x %d\n", heatmap.Size().Width, heatmap.Size().Height)
	img := image.NewRGBA(image.Rect(0, 0, int(heatmap.Size().Width), int(heatmap.Size().Height)))
	sum := uint(0)
	for x := uint(0); x < heatmap.Size().Width; x++ {
		for y := uint(0); y < heatmap.Size().Height; y++ {
			h := uint8(heatmap.At(x, y))
			sum += uint(h)
			img.Set(int(x), int(y), color.RGBA{h, h, h, math.MaxUint8})
		}
	}
	fmt.Printf("Average brightness %v\n", sum/(heatmap.Size().Width*heatmap.Size().Height))
	return img
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	heatmap := NewHeatmapBufferArray(300, 200)
	heatmap.Apply(spreadAction, 255)
	img := MakeImage(heatmap.Heatmap())
	fmt.Println(img.Bounds())
	fmt.Println(img.At(20, 20))
	f, err := os.Create("/Users/hoefel/Development/go/src/poster/image.png")
	check(err)
	w := bufio.NewWriter(f)
	check(png.Encode(w, img))
	check(w.Flush())
	check(f.Close())
}
