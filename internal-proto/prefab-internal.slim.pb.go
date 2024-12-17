// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.0
// 	protoc        v5.29.1
// source: prefab-internal.slim.proto

package prefab_internal_slim

import (
	proto "github.com/prefab-cloud/prefab-cloud-go/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ConfigWrapper struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Config        *proto.Config          `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	Deleted       bool                   `protobuf:"varint,2,opt,name=deleted,proto3" json:"deleted,omitempty"`
	CreatedAt     *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ConfigWrapper) Reset() {
	*x = ConfigWrapper{}
	mi := &file_prefab_internal_slim_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConfigWrapper) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfigWrapper) ProtoMessage() {}

func (x *ConfigWrapper) ProtoReflect() protoreflect.Message {
	mi := &file_prefab_internal_slim_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfigWrapper.ProtoReflect.Descriptor instead.
func (*ConfigWrapper) Descriptor() ([]byte, []int) {
	return file_prefab_internal_slim_proto_rawDescGZIP(), []int{0}
}

func (x *ConfigWrapper) GetConfig() *proto.Config {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *ConfigWrapper) GetDeleted() bool {
	if x != nil {
		return x.Deleted
	}
	return false
}

func (x *ConfigWrapper) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

type ConfigDump struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ProjectId     int64                  `protobuf:"varint,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	CreatedAt     *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	MaxConfigId   int64                  `protobuf:"varint,3,opt,name=max_config_id,json=maxConfigId,proto3" json:"max_config_id,omitempty"`
	Wrappers      []*ConfigWrapper       `protobuf:"bytes,4,rep,name=wrappers,proto3" json:"wrappers,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ConfigDump) Reset() {
	*x = ConfigDump{}
	mi := &file_prefab_internal_slim_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConfigDump) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfigDump) ProtoMessage() {}

func (x *ConfigDump) ProtoReflect() protoreflect.Message {
	mi := &file_prefab_internal_slim_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfigDump.ProtoReflect.Descriptor instead.
func (*ConfigDump) Descriptor() ([]byte, []int) {
	return file_prefab_internal_slim_proto_rawDescGZIP(), []int{1}
}

func (x *ConfigDump) GetProjectId() int64 {
	if x != nil {
		return x.ProjectId
	}
	return 0
}

func (x *ConfigDump) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *ConfigDump) GetMaxConfigId() int64 {
	if x != nil {
		return x.MaxConfigId
	}
	return 0
}

func (x *ConfigDump) GetWrappers() []*ConfigWrapper {
	if x != nil {
		return x.Wrappers
	}
	return nil
}

var File_prefab_internal_slim_proto protoreflect.FileDescriptor

var file_prefab_internal_slim_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x70, 0x72, 0x65, 0x66, 0x61, 0x62, 0x2d, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2e, 0x73, 0x6c, 0x69, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x70, 0x72,
	0x65, 0x66, 0x61, 0x62, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x73, 0x6c,
	0x69, 0x6d, 0x1a, 0x0c, 0x70, 0x72, 0x65, 0x66, 0x61, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x8c, 0x01, 0x0a, 0x0d, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x57, 0x72, 0x61, 0x70,
	0x70, 0x65, 0x72, 0x12, 0x26, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x72, 0x65, 0x66, 0x61, 0x62, 0x2e, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x18, 0x0a, 0x07, 0x64,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x64, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x64, 0x12, 0x39, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x5f, 0x61, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74,
	0x22, 0xcb, 0x01, 0x0a, 0x0a, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x44, 0x75, 0x6d, 0x70, 0x12,
	0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x39,
	0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x22, 0x0a, 0x0d, 0x6d, 0x61, 0x78,
	0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0b, 0x6d, 0x61, 0x78, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x49, 0x64, 0x12, 0x3f, 0x0a,
	0x08, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x23, 0x2e, 0x70, 0x72, 0x65, 0x66, 0x61, 0x62, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2e, 0x73, 0x6c, 0x69, 0x6d, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x57, 0x72, 0x61,
	0x70, 0x70, 0x65, 0x72, 0x52, 0x08, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x42, 0x16,
	0x5a, 0x14, 0x70, 0x72, 0x65, 0x66, 0x61, 0x62, 0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x5f, 0x73, 0x6c, 0x69, 0x6d, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_prefab_internal_slim_proto_rawDescOnce sync.Once
	file_prefab_internal_slim_proto_rawDescData = file_prefab_internal_slim_proto_rawDesc
)

func file_prefab_internal_slim_proto_rawDescGZIP() []byte {
	file_prefab_internal_slim_proto_rawDescOnce.Do(func() {
		file_prefab_internal_slim_proto_rawDescData = protoimpl.X.CompressGZIP(file_prefab_internal_slim_proto_rawDescData)
	})
	return file_prefab_internal_slim_proto_rawDescData
}

var file_prefab_internal_slim_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_prefab_internal_slim_proto_goTypes = []any{
	(*ConfigWrapper)(nil),         // 0: prefab.internal.slim.ConfigWrapper
	(*ConfigDump)(nil),            // 1: prefab.internal.slim.ConfigDump
	(*proto.Config)(nil),          // 2: prefab.Config
	(*timestamppb.Timestamp)(nil), // 3: google.protobuf.Timestamp
}
var file_prefab_internal_slim_proto_depIdxs = []int32{
	2, // 0: prefab.internal.slim.ConfigWrapper.config:type_name -> prefab.Config
	3, // 1: prefab.internal.slim.ConfigWrapper.created_at:type_name -> google.protobuf.Timestamp
	3, // 2: prefab.internal.slim.ConfigDump.created_at:type_name -> google.protobuf.Timestamp
	0, // 3: prefab.internal.slim.ConfigDump.wrappers:type_name -> prefab.internal.slim.ConfigWrapper
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_prefab_internal_slim_proto_init() }
func file_prefab_internal_slim_proto_init() {
	if File_prefab_internal_slim_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_prefab_internal_slim_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_prefab_internal_slim_proto_goTypes,
		DependencyIndexes: file_prefab_internal_slim_proto_depIdxs,
		MessageInfos:      file_prefab_internal_slim_proto_msgTypes,
	}.Build()
	File_prefab_internal_slim_proto = out.File
	file_prefab_internal_slim_proto_rawDesc = nil
	file_prefab_internal_slim_proto_goTypes = nil
	file_prefab_internal_slim_proto_depIdxs = nil
}
