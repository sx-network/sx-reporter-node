package proto

import "google.golang.org/protobuf/runtime/protoimpl"

type Report struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// marketHash is the market hash of the repot
	MarketHash string `protobuf:"bytes,1,opt,name=marketHash,proto3" json:"marketHash,omitempty"`
	// outcome is the outcome of the report
	Outcome int32 `protobuf:"varint,2,opt,name=outcome,proto3" json:"outcome,omitempty"`
}

type UnimplementedDataFeedOperatorServer struct {
}
