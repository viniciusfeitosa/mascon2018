# Running GRPC

From the V4 root diretory, run the command:

```sh
$ protoc -I protofiles .\protofiles\preferences.proto --go_out=plugins=grpc:preferences
```
