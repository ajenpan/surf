## gen proto

protoc -I./nnproto ./nnproto/*.proto --go_out=./nnproto_go/  --micro_out=./nnproto_go/ 