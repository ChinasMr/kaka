// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: kaka/v1/kaka.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// KakaClient is the client API for Kaka service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KakaClient interface {
	Debug(ctx context.Context, in *DebugRequest, opts ...grpc.CallOption) (*DebugReply, error)
}

type kakaClient struct {
	cc grpc.ClientConnInterface
}

func NewKakaClient(cc grpc.ClientConnInterface) KakaClient {
	return &kakaClient{cc}
}

func (c *kakaClient) Debug(ctx context.Context, in *DebugRequest, opts ...grpc.CallOption) (*DebugReply, error) {
	out := new(DebugReply)
	err := c.cc.Invoke(ctx, "/api.kaka.v1.Kaka/Debug", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KakaServer is the server API for Kaka service.
// All implementations must embed UnimplementedKakaServer
// for forward compatibility
type KakaServer interface {
	Debug(context.Context, *DebugRequest) (*DebugReply, error)
	mustEmbedUnimplementedKakaServer()
}

// UnimplementedKakaServer must be embedded to have forward compatible implementations.
type UnimplementedKakaServer struct {
}

func (UnimplementedKakaServer) Debug(context.Context, *DebugRequest) (*DebugReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Debug not implemented")
}
func (UnimplementedKakaServer) mustEmbedUnimplementedKakaServer() {}

// UnsafeKakaServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KakaServer will
// result in compilation errors.
type UnsafeKakaServer interface {
	mustEmbedUnimplementedKakaServer()
}

func RegisterKakaServer(s grpc.ServiceRegistrar, srv KakaServer) {
	s.RegisterService(&Kaka_ServiceDesc, srv)
}

func _Kaka_Debug_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DebugRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KakaServer).Debug(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.kaka.v1.Kaka/Debug",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KakaServer).Debug(ctx, req.(*DebugRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Kaka_ServiceDesc is the grpc.ServiceDesc for Kaka service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Kaka_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.kaka.v1.Kaka",
	HandlerType: (*KakaServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Debug",
			Handler:    _Kaka_Debug_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kaka/v1/kaka.proto",
}
