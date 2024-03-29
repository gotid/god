// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.9
// source: portal.proto

package portal

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

type PortalRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *PortalRequest) Reset() {
	*x = PortalRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_portal_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PortalRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortalRequest) ProtoMessage() {}

func (x *PortalRequest) ProtoReflect() protoreflect.Message {
	mi := &file_portal_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortalRequest.ProtoReflect.Descriptor instead.
func (*PortalRequest) Descriptor() ([]byte, []int) {
	return file_portal_proto_rawDescGZIP(), []int{0}
}

func (x *PortalRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type PortalResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Response string `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
}

func (x *PortalResponse) Reset() {
	*x = PortalResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_portal_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PortalResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortalResponse) ProtoMessage() {}

func (x *PortalResponse) ProtoReflect() protoreflect.Message {
	mi := &file_portal_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortalResponse.ProtoReflect.Descriptor instead.
func (*PortalResponse) Descriptor() ([]byte, []int) {
	return file_portal_proto_rawDescGZIP(), []int{1}
}

func (x *PortalResponse) GetResponse() string {
	if x != nil {
		return x.Response
	}
	return ""
}

var File_portal_proto protoreflect.FileDescriptor

var file_portal_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x22, 0x23, 0x0a, 0x0d, 0x50, 0x6f, 0x72, 0x74, 0x61, 0x6c,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x2c, 0x0a, 0x0e, 0x50,
	0x6f, 0x72, 0x74, 0x61, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1a, 0x0a,
	0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x41, 0x0a, 0x06, 0x50, 0x6f, 0x72,
	0x74, 0x61, 0x6c, 0x12, 0x37, 0x0a, 0x06, 0x50, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x12, 0x15, 0x2e,
	0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x2e, 0x50, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x2e, 0x50, 0x6f,
	0x72, 0x74, 0x61, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x0a, 0x5a, 0x08,
	0x2e, 0x2f, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_portal_proto_rawDescOnce sync.Once
	file_portal_proto_rawDescData = file_portal_proto_rawDesc
)

func file_portal_proto_rawDescGZIP() []byte {
	file_portal_proto_rawDescOnce.Do(func() {
		file_portal_proto_rawDescData = protoimpl.X.CompressGZIP(file_portal_proto_rawDescData)
	})
	return file_portal_proto_rawDescData
}

var file_portal_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_portal_proto_goTypes = []interface{}{
	(*PortalRequest)(nil),  // 0: portal.PortalRequest
	(*PortalResponse)(nil), // 1: portal.PortalResponse
}
var file_portal_proto_depIdxs = []int32{
	0, // 0: portal.Portal.Portal:input_type -> portal.PortalRequest
	1, // 1: portal.Portal.Portal:output_type -> portal.PortalResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_portal_proto_init() }
func file_portal_proto_init() {
	if File_portal_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_portal_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PortalRequest); i {
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
		file_portal_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PortalResponse); i {
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
			RawDescriptor: file_portal_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_portal_proto_goTypes,
		DependencyIndexes: file_portal_proto_depIdxs,
		MessageInfos:      file_portal_proto_msgTypes,
	}.Build()
	File_portal_proto = out.File
	file_portal_proto_rawDesc = nil
	file_portal_proto_goTypes = nil
	file_portal_proto_depIdxs = nil
}
