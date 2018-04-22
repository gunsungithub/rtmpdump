package main

/*
#cgo LDFLAGS: -lavutil -lavformat -lswscale -lswresample -lavcodec -lm
#include "stdlib.h"
#include "demo_ffmpeg.h"
*/
import "C"
import (
	"log"
	"math"
	"reflect"
	"unsafe"
)

type Slice struct {
	Data []byte
	data *c_slice_t
}

type Slice16 struct {
	Data []uint16
	data *c_slice_t
}

type c_slice_t struct {
	p unsafe.Pointer
	n int
}

var frame_index int = 0

//export fill_image_bytes
func fill_image_bytes() {

}

//export fill_image_bytes_GO
func fill_image_bytes_GO(Y, Cb, Cr unsafe.Pointer, width, height int) {
	var x, y int
	var i int
	i = frame_index
	frame_index++
	dataY := &c_slice_t{Y, width * height}
	sY := &Slice{data: dataY}
	hY := (*reflect.SliceHeader)((unsafe.Pointer(&sY.Data)))
	hY.Cap = dataY.n
	hY.Len = dataY.n
	hY.Data = uintptr(Y)
	/* Y */
	for y = 0; y < height; y++ {
		for x = 0; x < width; x++ {
			sY.Data[y*width+x] = (byte)(x + y + i*3)
		}
	}
	dataCb := &c_slice_t{Cb, width * height / 4}
	sCb := &Slice{data: dataCb}
	hCb := (*reflect.SliceHeader)((unsafe.Pointer(&sCb.Data)))
	hCb.Cap = dataCb.n
	hCb.Len = dataCb.n
	hCb.Data = uintptr(Cb)

	dataCr := &c_slice_t{Cr, width * height / 4}
	sCr := &Slice{data: dataCb}
	hCr := (*reflect.SliceHeader)((unsafe.Pointer(&sCr.Data)))
	hCr.Cap = dataCr.n
	hCr.Len = dataCr.n
	hCr.Data = uintptr(Cr)
	/* Cb and Cr */
	for y = 0; y < height/2; y++ {
		for x = 0; x < width/2; x++ {
			sCb.Data[y*(width>>1)+x] = (byte)(128 + y + i*2)
			sCr.Data[y*(width>>1)+x] = (byte)(64 + x + i*5)
		}
	}
}

var t float64 = 0
var tincr float64 = 2 * 3.141592653 * 110.0 / 44100
var tincr2 float64 = 2 * 3.141592653 * 110.0 / 44100 / 44100

//export fill_audio_bytes
func fill_audio_bytes() {

}

var once bool = true

//export fill_audio_bytes_GO
func fill_audio_bytes_GO(buf unsafe.Pointer, nb_samples, channels int) {
	var j, i int
	var v uint16
	data := &c_slice_t{buf, nb_samples * channels * 2}
	s := &Slice16{data: data}
	h := (*reflect.SliceHeader)((unsafe.Pointer(&s.Data)))
	h.Cap = data.n
	h.Len = data.n
	h.Data = uintptr(buf)
	if once {
		log.Println(nb_samples, channels, buf, (&(s.Data[data.n-2])), (nb_samples-1)*nb_samples+(channels-1)*channels)
		once = false
	}

	for j = 0; j < nb_samples; j++ {
		v = (uint16)(math.Sin(t) * 10000)
		for i = 0; i < channels; i++ {
			s.Data[j*channels+i*2] = v
		}
		t += tincr
		tincr += tincr2
	}
}

func main() {
	log.Printf("hello world")
	//video_cb := syscall.NewCallback(fill_image_bytes)
	//C.set_video_callback(video_cb)
	//audio_cb := syscall.NewCallback(fill_audio_bytes)
	//C.set_video_callback(audio_cb)
	cc := C.CString("rtmp://25582.lsspublish.aodianyun.com/server0/stream")
	C.setup(cc)
	C.free(unsafe.Pointer(cc))
	return
}
