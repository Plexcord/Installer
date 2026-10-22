//go:build !cli

package main

import (
	"image/color"
	"syscall"
	"unsafe"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// StyleTitleBar paints the native title bar in the theme colours instead of drawing a custom one,
// so dragging, snapping and the window buttons keep working. Colours need Windows 11; Windows 10 only gets the dark bar.
func StyleTitleBar() {
	w := glfw.GetCurrentContext()
	set := syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")
	if w == nil || set.Find() != nil {
		return
	}
	hwnd := uintptr(unsafe.Pointer(w.GetWin32Window()))

	for _, a := range [][2]uint32{
		{20, 1},             // DWMWA_USE_IMMERSIVE_DARK_MODE
		{34, colorRef(Bg)},  // DWMWA_BORDER_COLOR
		{35, colorRef(Bg)},  // DWMWA_CAPTION_COLOR
		{36, colorRef(Ink)}, // DWMWA_TEXT_COLOR
	} {
		v := a[1]
		_, _, _ = set.Call(hwnd, uintptr(a[0]), uintptr(unsafe.Pointer(&v)), unsafe.Sizeof(v))
	}
}

// colorRef converts to the 0x00BBGGRR layout Windows expects
func colorRef(c color.RGBA) uint32 {
	return uint32(c.B)<<16 | uint32(c.G)<<8 | uint32(c.R)
}
