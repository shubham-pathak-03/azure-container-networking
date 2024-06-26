// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v3.12.4
// source: cns/grpc/proto/server.proto

package v1alpha

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// SetOrchestratorInfoRequest is the request message for setting the orchestrator information.
type SetOrchestratorInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DncPartitionKey  string `protobuf:"bytes,1,opt,name=dncPartitionKey,proto3" json:"dncPartitionKey,omitempty"`   // The partition key for DNC.
	NodeID           string `protobuf:"bytes,2,opt,name=nodeID,proto3" json:"nodeID,omitempty"`                     // The node ID.
	OrchestratorType string `protobuf:"bytes,3,opt,name=orchestratorType,proto3" json:"orchestratorType,omitempty"` // The type of the orchestrator.
}

func (x *SetOrchestratorInfoRequest) Reset() {
	*x = SetOrchestratorInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cns_grpc_proto_server_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetOrchestratorInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetOrchestratorInfoRequest) ProtoMessage() {}

func (x *SetOrchestratorInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_cns_grpc_proto_server_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetOrchestratorInfoRequest.ProtoReflect.Descriptor instead.
func (*SetOrchestratorInfoRequest) Descriptor() ([]byte, []int) {
	return file_cns_grpc_proto_server_proto_rawDescGZIP(), []int{0}
}

func (x *SetOrchestratorInfoRequest) GetDncPartitionKey() string {
	if x != nil {
		return x.DncPartitionKey
	}
	return ""
}

func (x *SetOrchestratorInfoRequest) GetNodeID() string {
	if x != nil {
		return x.NodeID
	}
	return ""
}

func (x *SetOrchestratorInfoRequest) GetOrchestratorType() string {
	if x != nil {
		return x.OrchestratorType
	}
	return ""
}

// SetOrchestratorInfoResponse is the response message for setting the orchestrator information.
type SetOrchestratorInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SetOrchestratorInfoResponse) Reset() {
	*x = SetOrchestratorInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cns_grpc_proto_server_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetOrchestratorInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetOrchestratorInfoResponse) ProtoMessage() {}

func (x *SetOrchestratorInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cns_grpc_proto_server_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetOrchestratorInfoResponse.ProtoReflect.Descriptor instead.
func (*SetOrchestratorInfoResponse) Descriptor() ([]byte, []int) {
	return file_cns_grpc_proto_server_proto_rawDescGZIP(), []int{1}
}

// NodeInfoRequest is the request message for retrieving detailed information about a specific node.
type NodeInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeID string `protobuf:"bytes,1,opt,name=nodeID,proto3" json:"nodeID,omitempty"` // The node ID to identify the specific node.
}

func (x *NodeInfoRequest) Reset() {
	*x = NodeInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cns_grpc_proto_server_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeInfoRequest) ProtoMessage() {}

func (x *NodeInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_cns_grpc_proto_server_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeInfoRequest.ProtoReflect.Descriptor instead.
func (*NodeInfoRequest) Descriptor() ([]byte, []int) {
	return file_cns_grpc_proto_server_proto_rawDescGZIP(), []int{2}
}

func (x *NodeInfoRequest) GetNodeID() string {
	if x != nil {
		return x.NodeID
	}
	return ""
}

// NodeInfoResponse is the response message containing detailed information about a specific node.
type NodeInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeID    string `protobuf:"bytes,1,opt,name=nodeID,proto3" json:"nodeID,omitempty"`        // The node ID.
	Name      string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`            // The name of the node.
	Ip        string `protobuf:"bytes,3,opt,name=ip,proto3" json:"ip,omitempty"`                // The IP address of the node.
	IsHealthy bool   `protobuf:"varint,4,opt,name=isHealthy,proto3" json:"isHealthy,omitempty"` // Indicates whether the node is healthy or not.
	Status    string `protobuf:"bytes,5,opt,name=status,proto3" json:"status,omitempty"`        // The current status of the node (e.g., running, stopped).
	Message   string `protobuf:"bytes,6,opt,name=message,proto3" json:"message,omitempty"`      // Additional information about the node's health or status.
}

func (x *NodeInfoResponse) Reset() {
	*x = NodeInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cns_grpc_proto_server_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeInfoResponse) ProtoMessage() {}

func (x *NodeInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cns_grpc_proto_server_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeInfoResponse.ProtoReflect.Descriptor instead.
func (*NodeInfoResponse) Descriptor() ([]byte, []int) {
	return file_cns_grpc_proto_server_proto_rawDescGZIP(), []int{3}
}

func (x *NodeInfoResponse) GetNodeID() string {
	if x != nil {
		return x.NodeID
	}
	return ""
}

func (x *NodeInfoResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *NodeInfoResponse) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *NodeInfoResponse) GetIsHealthy() bool {
	if x != nil {
		return x.IsHealthy
	}
	return false
}

func (x *NodeInfoResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *NodeInfoResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_cns_grpc_proto_server_proto protoreflect.FileDescriptor

var file_cns_grpc_proto_server_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x63, 0x6e, 0x73, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x63,
	0x6e, 0x73, 0x22, 0x8a, 0x01, 0x0a, 0x1a, 0x53, 0x65, 0x74, 0x4f, 0x72, 0x63, 0x68, 0x65, 0x73,
	0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x28, 0x0a, 0x0f, 0x64, 0x6e, 0x63, 0x50, 0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x64, 0x6e, 0x63, 0x50,
	0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x4b, 0x65, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x6e,
	0x6f, 0x64, 0x65, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6e, 0x6f, 0x64,
	0x65, 0x49, 0x44, 0x12, 0x2a, 0x0a, 0x10, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61,
	0x74, 0x6f, 0x72, 0x54, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x6f,
	0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x54, 0x79, 0x70, 0x65, 0x22,
	0x1d, 0x0a, 0x1b, 0x53, 0x65, 0x74, 0x4f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74,
	0x6f, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x29,
	0x0a, 0x0f, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x16, 0x0a, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x44, 0x22, 0x9e, 0x01, 0x0a, 0x10, 0x4e, 0x6f,
	0x64, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x6e, 0x6f, 0x64, 0x65, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x73,
	0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x69,
	0x73, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0x9b, 0x01, 0x0a, 0x03, 0x43,
	0x4e, 0x53, 0x12, 0x58, 0x0a, 0x13, 0x53, 0x65, 0x74, 0x4f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1f, 0x2e, 0x63, 0x6e, 0x73, 0x2e,
	0x53, 0x65, 0x74, 0x4f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x63, 0x6e, 0x73,
	0x2e, 0x53, 0x65, 0x74, 0x4f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3a, 0x0a, 0x0b,
	0x47, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x14, 0x2e, 0x63, 0x6e,
	0x73, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x15, 0x2e, 0x63, 0x6e, 0x73, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x6e, 0x66, 0x6f,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x12, 0x5a, 0x10, 0x63, 0x6e, 0x73, 0x2f,
	0x67, 0x72, 0x70, 0x63, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cns_grpc_proto_server_proto_rawDescOnce sync.Once
	file_cns_grpc_proto_server_proto_rawDescData = file_cns_grpc_proto_server_proto_rawDesc
)

func file_cns_grpc_proto_server_proto_rawDescGZIP() []byte {
	file_cns_grpc_proto_server_proto_rawDescOnce.Do(func() {
		file_cns_grpc_proto_server_proto_rawDescData = protoimpl.X.CompressGZIP(file_cns_grpc_proto_server_proto_rawDescData)
	})
	return file_cns_grpc_proto_server_proto_rawDescData
}

var file_cns_grpc_proto_server_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_cns_grpc_proto_server_proto_goTypes = []interface{}{
	(*SetOrchestratorInfoRequest)(nil),  // 0: cns.SetOrchestratorInfoRequest
	(*SetOrchestratorInfoResponse)(nil), // 1: cns.SetOrchestratorInfoResponse
	(*NodeInfoRequest)(nil),             // 2: cns.NodeInfoRequest
	(*NodeInfoResponse)(nil),            // 3: cns.NodeInfoResponse
}
var file_cns_grpc_proto_server_proto_depIdxs = []int32{
	0, // 0: cns.CNS.SetOrchestratorInfo:input_type -> cns.SetOrchestratorInfoRequest
	2, // 1: cns.CNS.GetNodeInfo:input_type -> cns.NodeInfoRequest
	1, // 2: cns.CNS.SetOrchestratorInfo:output_type -> cns.SetOrchestratorInfoResponse
	3, // 3: cns.CNS.GetNodeInfo:output_type -> cns.NodeInfoResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_cns_grpc_proto_server_proto_init() }
func file_cns_grpc_proto_server_proto_init() {
	if File_cns_grpc_proto_server_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cns_grpc_proto_server_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetOrchestratorInfoRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_cns_grpc_proto_server_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetOrchestratorInfoResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_cns_grpc_proto_server_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeInfoRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_cns_grpc_proto_server_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeInfoResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_cns_grpc_proto_server_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_cns_grpc_proto_server_proto_goTypes,
		DependencyIndexes: file_cns_grpc_proto_server_proto_depIdxs,
		MessageInfos:      file_cns_grpc_proto_server_proto_msgTypes,
	}.Build()
	File_cns_grpc_proto_server_proto = out.File
	file_cns_grpc_proto_server_proto_rawDesc = nil
	file_cns_grpc_proto_server_proto_goTypes = nil
	file_cns_grpc_proto_server_proto_depIdxs = nil
}
