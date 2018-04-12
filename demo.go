package main

/*
#cgo LDFLAGS: -Llibrtmp -lrtmp
#include <stdlib.h>
#include "rtmp_sample_api.h"
#include <sys/times.h>
#include <unistd.h>
*/
import "C"
import (
	"bufio"
	"log"
	"os"
	"time"
	"unsafe"
)

var clk_tck C.long = 0

func GetTime() uint {
	var t C.struct_tms
	if clk_tck == 0 {
		clk_tck = C.sysconf(C._SC_CLK_TCK)
		log.Println("Hello", clk_tck)
	}
	return uint(C.times(&t) * 1000 / C.ulong(clk_tck))
}
func main() {
	log.Println("Hello, World!")
	f, err := os.OpenFile("test.flv", os.O_RDONLY, 0666)
	if err != nil {
		log.Println("OpenFile fail!")
		return
	}
	defer f.Close()
	C.rtmp_sample_init()
	defer C.rtmp_sample_final()
	url := C.CString("rtmp://25582.lsspublish.aodianyun.com/server0/stream")
	defer C.free(unsafe.Pointer(url))
	C.rtmp_sample_connect(url)
	defer C.rtmp_sample_disconnect()
	//f.Seek(9, os.SEEK_SET)
	//f.Seek(4, os.SEEK_CUR)
	start_time := GetTime()
	var pre_frame_time uint = 0
	var lasttime uint = 0
	bNextIsKey := false
	i := 0
	var datalength uint = 0
	var timestamp uint = 0
	reader := bufio.NewReaderSize(f, 1024*1024)
	reader.Discard(13)
	for i < 1000 {

		//log.Println("i:", i, bNextIsKey, GetTime(), start_time, pre_frame_time)
		if GetTime()-start_time < pre_frame_time && bNextIsKey {
			if pre_frame_time > lasttime {
				log.Println("Time Stamp:", pre_frame_time, "ms", i)
				lasttime = pre_frame_time
			}
			//i++
			time.Sleep(time.Second)
			continue
		}
		i++
		//log.Println("i:", i)
		//f.Seek(1, os.SEEK_CUR)

		buf, _ := reader.Peek(8)
		datalength = uint(buf[1]) << 16
		//log.Println("tmp:", buf[1])
		//tmp, _ = reader.ReadByte()
		datalength += uint(buf[2]) << 8
		//log.Println("tmp:", buf[2])
		//tmp, _ = reader.ReadByte()
		datalength += uint(buf[3])
		//log.Println("tmp:", buf[3])
		//log.Println("datalength:", datalength)
		//tmp, _ = reader.ReadByte()
		timestamp = 0
		timestamp = uint(buf[4]) << 16
		//tmp, _ = reader.ReadByte()
		timestamp += uint(buf[5]) << 8
		//tmp, _ = reader.ReadByte()
		timestamp += uint(buf[6])
		//tmp, _ = reader.ReadByte()
		//timestamp += uint(buf[7])
		//log.Println("timestamp:", timestamp, buf[4:8])
		//f.Seek(-1, os.SEEK_CUR)
		buf, err := reader.Peek(int(datalength) + 15)
		if err != nil {
			log.Println(err, datalength)
			reader.Discard(int(datalength) + 15)
			continue
		}
		//log.Println("buf:", buf[:5])
		reader.Discard(int(datalength) + 15)
		//f.Seek(int64(datalength)+15, os.SEEK_CUR)
		cc := C.CString(string(buf))
		pre_frame_time = timestamp
		C.rtmp_sample_add_data(cc, C.int(datalength)+15)
		C.free(unsafe.Pointer(cc))
		buf, _ = reader.Peek(12)
		if buf[0] == 0x09 {
			//f.Seek(11, os.SEEK_CUR)
			//buf, _ = reader.Peek(1)
			if buf[11] == 0x17 {
				bNextIsKey = true
			} else {
				bNextIsKey = false
			}
			//f.Seek(-11, os.SEEK_CUR)
		}
	}
}
