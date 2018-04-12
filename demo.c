#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <unistd.h>
#include <sys/times.h>

#include "rtmp_sample_api.h"
#include "demo.h"

#define HTON16(x)  ((x>>8&0xff)|(x<<8&0xff00))
#define HTON24(x)  ((x>>16&0xff)|(x<<16&0xff0000)|(x&0xff00))
#define HTON32(x)  ((x>>24&0xff)|(x>>8&0xff00)|\
        (x<<8&0xff0000)|(x<<24&0xff000000))
#define HTONTIME(x) ((x>>16&0xff)|(x<<16&0xff0000)|(x&0xff00)|(x&0xff000000))

#if !defined(RTMP_SERVER_URL) || !defined(LOCAL_FILE)
#define RTMP_SERVER_URL     "rtmp address"
#define LOCAL_FILE         "test flv file name"
#endif

/*read 1 byte*/
int ReadU8(uint32_t *u8,FILE*fp){
    if(fread(u8,1,1,fp)!=1)
        return 0;
    return 1;
}
/*read 2 byte*/
int ReadU16(uint32_t *u16,FILE*fp){
    if(fread(u16,2,1,fp)!=1)
        return 0;
    *u16=HTON16(*u16);
    return 1;
}
/*read 3 byte*/
int ReadU24(uint32_t *u24,FILE*fp){
    if(fread(u24,3,1,fp)!=1)
        return 0;
    *u24=HTON24(*u24);
    return 1;
}
/*read 4 byte*/
int ReadU32(uint32_t *u32,FILE*fp){
    if(fread(u32,4,1,fp)!=1)
        return 0;
    *u32=HTON32(*u32);
    return 1;
}
/*read 1 byte,and loopback 1 byte at once*/
int PeekU8(uint32_t *u8,FILE*fp){
    if(fread(u8,1,1,fp)!=1)
        return 0;
    fseek(fp,-1,SEEK_CUR);
    return 1;
}
/*read 4 byte and convert to time format*/
int ReadTime(uint32_t *utime,FILE*fp){
    if(fread(utime,4,1,fp)!=1)
        return 0;
    *utime=HTONTIME(*utime);
    return 1;
}
static int clk_tck;
uint32_t GetTime()
{
  struct tms t;
  if (!clk_tck) clk_tck = sysconf(_SC_CLK_TCK);
  return times(&t) * 1000 / clk_tck;
}
//Publish using RTMP_Write()
int publish_using_write(){
    uint32_t start_time=0;
    uint32_t now_time=0;
    uint32_t pre_frame_time=0;
    uint32_t lasttime=0;
    int bNextIsKey=0;
    char* pFileBuf=NULL;

    //read from tag header
    uint32_t type=0;
    uint32_t datalength=0;
    uint32_t timestamp=0;

    FILE*fp=NULL;
    fp=fopen(LOCAL_FILE,"rb");
    if (!fp){
        printf("Open File Error.\n");
        return -1;
    }

    /* set log level */
    //RTMP_LogLevel loglvl=RTMP_LOGDEBUG;
    //RTMP_LogSetLevel(loglvl);
    rtmp_sample_init();
    rtmp_sample_connect(RTMP_SERVER_URL);

    printf("Start to send data ...\n");
    //jump over FLV Header
    fseek(fp,9,SEEK_SET);
    //jump over previousTagSizen
    fseek(fp,4,SEEK_CUR);
    start_time=GetTime();
    while(1)
    {
        if((((now_time=GetTime())-start_time)<(pre_frame_time)) && bNextIsKey){
            //wait for 1 sec if the send process is too fast
            //this mechanism is not very good,need some improvement
            if(pre_frame_time>lasttime){
                printf("TimeStamp:%8u ms\n",pre_frame_time);
                lasttime=pre_frame_time;
            }
            sleep(1);
            continue;
        }

        //jump over type
        fseek(fp,1,SEEK_CUR);
        if(!ReadU24(&datalength,fp))
            break;
        if(!ReadTime(&timestamp,fp))
            break;
        //jump back
        fseek(fp,-8,SEEK_CUR);

        pFileBuf=(char*)malloc(11+datalength+4);
        memset(pFileBuf,0,11+datalength+4);
        if(fread(pFileBuf,1,11+datalength+4,fp)!=(11+datalength+4))
            break;

        pre_frame_time=timestamp;
        rtmp_sample_add_data(pFileBuf, 11 + datalength + 4);

        free(pFileBuf);
        pFileBuf=NULL;

        if(!PeekU8(&type,fp))
            break;
        if(type==0x09){
            if(fseek(fp,11,SEEK_CUR)!=0)
                break;
            if(!PeekU8(&type,fp)){
                break;
            }
            if(type==0x17)
                bNextIsKey=1;
            else
                bNextIsKey=0;
            fseek(fp,-11,SEEK_CUR);
        }
    }

    printf("\nSend Data Over\n");

    if(fp)
        fclose(fp);

    rtmp_sample_disconnect();
    rtmp_sample_final();

    if(pFileBuf){
        free(pFileBuf);
        pFileBuf=NULL;
    }

    return 0;
}

int main(int argc, char* argv[]){
    publish_using_write();
    return 0;
}
