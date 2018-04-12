int rtmp_sample_init();
int rtmp_sample_connect(char *url);
int rtmp_sample_add_data(char *pbuf, int datalength);
int rtmp_sample_disconnect();
int rtmp_sample_final();