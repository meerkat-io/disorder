rpc packet

request

| packet_size(4 bytes) | headers(map[string]string) | service_name(short string) | method_name(short string) | body(any) |

response

| packet_size(4 bytes) | status_code(byte) | body(any)/error(string) |