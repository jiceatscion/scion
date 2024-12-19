// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: proto/control_plane/v1/drkey.proto

package control_planeconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	control_plane "github.com/scionproto/scion/pkg/proto/control_plane"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// DRKeyInterServiceName is the fully-qualified name of the DRKeyInterService service.
	DRKeyInterServiceName = "proto.control_plane.v1.DRKeyInterService"
	// DRKeyIntraServiceName is the fully-qualified name of the DRKeyIntraService service.
	DRKeyIntraServiceName = "proto.control_plane.v1.DRKeyIntraService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// DRKeyInterServiceDRKeyLevel1Procedure is the fully-qualified name of the DRKeyInterService's
	// DRKeyLevel1 RPC.
	DRKeyInterServiceDRKeyLevel1Procedure = "/proto.control_plane.v1.DRKeyInterService/DRKeyLevel1"
	// DRKeyIntraServiceDRKeyIntraLevel1Procedure is the fully-qualified name of the DRKeyIntraService's
	// DRKeyIntraLevel1 RPC.
	DRKeyIntraServiceDRKeyIntraLevel1Procedure = "/proto.control_plane.v1.DRKeyIntraService/DRKeyIntraLevel1"
	// DRKeyIntraServiceDRKeyASHostProcedure is the fully-qualified name of the DRKeyIntraService's
	// DRKeyASHost RPC.
	DRKeyIntraServiceDRKeyASHostProcedure = "/proto.control_plane.v1.DRKeyIntraService/DRKeyASHost"
	// DRKeyIntraServiceDRKeyHostASProcedure is the fully-qualified name of the DRKeyIntraService's
	// DRKeyHostAS RPC.
	DRKeyIntraServiceDRKeyHostASProcedure = "/proto.control_plane.v1.DRKeyIntraService/DRKeyHostAS"
	// DRKeyIntraServiceDRKeyHostHostProcedure is the fully-qualified name of the DRKeyIntraService's
	// DRKeyHostHost RPC.
	DRKeyIntraServiceDRKeyHostHostProcedure = "/proto.control_plane.v1.DRKeyIntraService/DRKeyHostHost"
	// DRKeyIntraServiceDRKeySecretValueProcedure is the fully-qualified name of the DRKeyIntraService's
	// DRKeySecretValue RPC.
	DRKeyIntraServiceDRKeySecretValueProcedure = "/proto.control_plane.v1.DRKeyIntraService/DRKeySecretValue"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	dRKeyInterServiceServiceDescriptor                = control_plane.File_proto_control_plane_v1_drkey_proto.Services().ByName("DRKeyInterService")
	dRKeyInterServiceDRKeyLevel1MethodDescriptor      = dRKeyInterServiceServiceDescriptor.Methods().ByName("DRKeyLevel1")
	dRKeyIntraServiceServiceDescriptor                = control_plane.File_proto_control_plane_v1_drkey_proto.Services().ByName("DRKeyIntraService")
	dRKeyIntraServiceDRKeyIntraLevel1MethodDescriptor = dRKeyIntraServiceServiceDescriptor.Methods().ByName("DRKeyIntraLevel1")
	dRKeyIntraServiceDRKeyASHostMethodDescriptor      = dRKeyIntraServiceServiceDescriptor.Methods().ByName("DRKeyASHost")
	dRKeyIntraServiceDRKeyHostASMethodDescriptor      = dRKeyIntraServiceServiceDescriptor.Methods().ByName("DRKeyHostAS")
	dRKeyIntraServiceDRKeyHostHostMethodDescriptor    = dRKeyIntraServiceServiceDescriptor.Methods().ByName("DRKeyHostHost")
	dRKeyIntraServiceDRKeySecretValueMethodDescriptor = dRKeyIntraServiceServiceDescriptor.Methods().ByName("DRKeySecretValue")
)

// DRKeyInterServiceClient is a client for the proto.control_plane.v1.DRKeyInterService service.
type DRKeyInterServiceClient interface {
	DRKeyLevel1(context.Context, *connect.Request[control_plane.DRKeyLevel1Request]) (*connect.Response[control_plane.DRKeyLevel1Response], error)
}

// NewDRKeyInterServiceClient constructs a client for the proto.control_plane.v1.DRKeyInterService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewDRKeyInterServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) DRKeyInterServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &dRKeyInterServiceClient{
		dRKeyLevel1: connect.NewClient[control_plane.DRKeyLevel1Request, control_plane.DRKeyLevel1Response](
			httpClient,
			baseURL+DRKeyInterServiceDRKeyLevel1Procedure,
			connect.WithSchema(dRKeyInterServiceDRKeyLevel1MethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// dRKeyInterServiceClient implements DRKeyInterServiceClient.
type dRKeyInterServiceClient struct {
	dRKeyLevel1 *connect.Client[control_plane.DRKeyLevel1Request, control_plane.DRKeyLevel1Response]
}

// DRKeyLevel1 calls proto.control_plane.v1.DRKeyInterService.DRKeyLevel1.
func (c *dRKeyInterServiceClient) DRKeyLevel1(ctx context.Context, req *connect.Request[control_plane.DRKeyLevel1Request]) (*connect.Response[control_plane.DRKeyLevel1Response], error) {
	return c.dRKeyLevel1.CallUnary(ctx, req)
}

// DRKeyInterServiceHandler is an implementation of the proto.control_plane.v1.DRKeyInterService
// service.
type DRKeyInterServiceHandler interface {
	DRKeyLevel1(context.Context, *connect.Request[control_plane.DRKeyLevel1Request]) (*connect.Response[control_plane.DRKeyLevel1Response], error)
}

// NewDRKeyInterServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewDRKeyInterServiceHandler(svc DRKeyInterServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	dRKeyInterServiceDRKeyLevel1Handler := connect.NewUnaryHandler(
		DRKeyInterServiceDRKeyLevel1Procedure,
		svc.DRKeyLevel1,
		connect.WithSchema(dRKeyInterServiceDRKeyLevel1MethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/proto.control_plane.v1.DRKeyInterService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case DRKeyInterServiceDRKeyLevel1Procedure:
			dRKeyInterServiceDRKeyLevel1Handler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedDRKeyInterServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedDRKeyInterServiceHandler struct{}

func (UnimplementedDRKeyInterServiceHandler) DRKeyLevel1(context.Context, *connect.Request[control_plane.DRKeyLevel1Request]) (*connect.Response[control_plane.DRKeyLevel1Response], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("proto.control_plane.v1.DRKeyInterService.DRKeyLevel1 is not implemented"))
}

// DRKeyIntraServiceClient is a client for the proto.control_plane.v1.DRKeyIntraService service.
type DRKeyIntraServiceClient interface {
	DRKeyIntraLevel1(context.Context, *connect.Request[control_plane.DRKeyIntraLevel1Request]) (*connect.Response[control_plane.DRKeyIntraLevel1Response], error)
	DRKeyASHost(context.Context, *connect.Request[control_plane.DRKeyASHostRequest]) (*connect.Response[control_plane.DRKeyASHostResponse], error)
	DRKeyHostAS(context.Context, *connect.Request[control_plane.DRKeyHostASRequest]) (*connect.Response[control_plane.DRKeyHostASResponse], error)
	DRKeyHostHost(context.Context, *connect.Request[control_plane.DRKeyHostHostRequest]) (*connect.Response[control_plane.DRKeyHostHostResponse], error)
	DRKeySecretValue(context.Context, *connect.Request[control_plane.DRKeySecretValueRequest]) (*connect.Response[control_plane.DRKeySecretValueResponse], error)
}

// NewDRKeyIntraServiceClient constructs a client for the proto.control_plane.v1.DRKeyIntraService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewDRKeyIntraServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) DRKeyIntraServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &dRKeyIntraServiceClient{
		dRKeyIntraLevel1: connect.NewClient[control_plane.DRKeyIntraLevel1Request, control_plane.DRKeyIntraLevel1Response](
			httpClient,
			baseURL+DRKeyIntraServiceDRKeyIntraLevel1Procedure,
			connect.WithSchema(dRKeyIntraServiceDRKeyIntraLevel1MethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		dRKeyASHost: connect.NewClient[control_plane.DRKeyASHostRequest, control_plane.DRKeyASHostResponse](
			httpClient,
			baseURL+DRKeyIntraServiceDRKeyASHostProcedure,
			connect.WithSchema(dRKeyIntraServiceDRKeyASHostMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		dRKeyHostAS: connect.NewClient[control_plane.DRKeyHostASRequest, control_plane.DRKeyHostASResponse](
			httpClient,
			baseURL+DRKeyIntraServiceDRKeyHostASProcedure,
			connect.WithSchema(dRKeyIntraServiceDRKeyHostASMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		dRKeyHostHost: connect.NewClient[control_plane.DRKeyHostHostRequest, control_plane.DRKeyHostHostResponse](
			httpClient,
			baseURL+DRKeyIntraServiceDRKeyHostHostProcedure,
			connect.WithSchema(dRKeyIntraServiceDRKeyHostHostMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		dRKeySecretValue: connect.NewClient[control_plane.DRKeySecretValueRequest, control_plane.DRKeySecretValueResponse](
			httpClient,
			baseURL+DRKeyIntraServiceDRKeySecretValueProcedure,
			connect.WithSchema(dRKeyIntraServiceDRKeySecretValueMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// dRKeyIntraServiceClient implements DRKeyIntraServiceClient.
type dRKeyIntraServiceClient struct {
	dRKeyIntraLevel1 *connect.Client[control_plane.DRKeyIntraLevel1Request, control_plane.DRKeyIntraLevel1Response]
	dRKeyASHost      *connect.Client[control_plane.DRKeyASHostRequest, control_plane.DRKeyASHostResponse]
	dRKeyHostAS      *connect.Client[control_plane.DRKeyHostASRequest, control_plane.DRKeyHostASResponse]
	dRKeyHostHost    *connect.Client[control_plane.DRKeyHostHostRequest, control_plane.DRKeyHostHostResponse]
	dRKeySecretValue *connect.Client[control_plane.DRKeySecretValueRequest, control_plane.DRKeySecretValueResponse]
}

// DRKeyIntraLevel1 calls proto.control_plane.v1.DRKeyIntraService.DRKeyIntraLevel1.
func (c *dRKeyIntraServiceClient) DRKeyIntraLevel1(ctx context.Context, req *connect.Request[control_plane.DRKeyIntraLevel1Request]) (*connect.Response[control_plane.DRKeyIntraLevel1Response], error) {
	return c.dRKeyIntraLevel1.CallUnary(ctx, req)
}

// DRKeyASHost calls proto.control_plane.v1.DRKeyIntraService.DRKeyASHost.
func (c *dRKeyIntraServiceClient) DRKeyASHost(ctx context.Context, req *connect.Request[control_plane.DRKeyASHostRequest]) (*connect.Response[control_plane.DRKeyASHostResponse], error) {
	return c.dRKeyASHost.CallUnary(ctx, req)
}

// DRKeyHostAS calls proto.control_plane.v1.DRKeyIntraService.DRKeyHostAS.
func (c *dRKeyIntraServiceClient) DRKeyHostAS(ctx context.Context, req *connect.Request[control_plane.DRKeyHostASRequest]) (*connect.Response[control_plane.DRKeyHostASResponse], error) {
	return c.dRKeyHostAS.CallUnary(ctx, req)
}

// DRKeyHostHost calls proto.control_plane.v1.DRKeyIntraService.DRKeyHostHost.
func (c *dRKeyIntraServiceClient) DRKeyHostHost(ctx context.Context, req *connect.Request[control_plane.DRKeyHostHostRequest]) (*connect.Response[control_plane.DRKeyHostHostResponse], error) {
	return c.dRKeyHostHost.CallUnary(ctx, req)
}

// DRKeySecretValue calls proto.control_plane.v1.DRKeyIntraService.DRKeySecretValue.
func (c *dRKeyIntraServiceClient) DRKeySecretValue(ctx context.Context, req *connect.Request[control_plane.DRKeySecretValueRequest]) (*connect.Response[control_plane.DRKeySecretValueResponse], error) {
	return c.dRKeySecretValue.CallUnary(ctx, req)
}

// DRKeyIntraServiceHandler is an implementation of the proto.control_plane.v1.DRKeyIntraService
// service.
type DRKeyIntraServiceHandler interface {
	DRKeyIntraLevel1(context.Context, *connect.Request[control_plane.DRKeyIntraLevel1Request]) (*connect.Response[control_plane.DRKeyIntraLevel1Response], error)
	DRKeyASHost(context.Context, *connect.Request[control_plane.DRKeyASHostRequest]) (*connect.Response[control_plane.DRKeyASHostResponse], error)
	DRKeyHostAS(context.Context, *connect.Request[control_plane.DRKeyHostASRequest]) (*connect.Response[control_plane.DRKeyHostASResponse], error)
	DRKeyHostHost(context.Context, *connect.Request[control_plane.DRKeyHostHostRequest]) (*connect.Response[control_plane.DRKeyHostHostResponse], error)
	DRKeySecretValue(context.Context, *connect.Request[control_plane.DRKeySecretValueRequest]) (*connect.Response[control_plane.DRKeySecretValueResponse], error)
}

// NewDRKeyIntraServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewDRKeyIntraServiceHandler(svc DRKeyIntraServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	dRKeyIntraServiceDRKeyIntraLevel1Handler := connect.NewUnaryHandler(
		DRKeyIntraServiceDRKeyIntraLevel1Procedure,
		svc.DRKeyIntraLevel1,
		connect.WithSchema(dRKeyIntraServiceDRKeyIntraLevel1MethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	dRKeyIntraServiceDRKeyASHostHandler := connect.NewUnaryHandler(
		DRKeyIntraServiceDRKeyASHostProcedure,
		svc.DRKeyASHost,
		connect.WithSchema(dRKeyIntraServiceDRKeyASHostMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	dRKeyIntraServiceDRKeyHostASHandler := connect.NewUnaryHandler(
		DRKeyIntraServiceDRKeyHostASProcedure,
		svc.DRKeyHostAS,
		connect.WithSchema(dRKeyIntraServiceDRKeyHostASMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	dRKeyIntraServiceDRKeyHostHostHandler := connect.NewUnaryHandler(
		DRKeyIntraServiceDRKeyHostHostProcedure,
		svc.DRKeyHostHost,
		connect.WithSchema(dRKeyIntraServiceDRKeyHostHostMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	dRKeyIntraServiceDRKeySecretValueHandler := connect.NewUnaryHandler(
		DRKeyIntraServiceDRKeySecretValueProcedure,
		svc.DRKeySecretValue,
		connect.WithSchema(dRKeyIntraServiceDRKeySecretValueMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/proto.control_plane.v1.DRKeyIntraService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case DRKeyIntraServiceDRKeyIntraLevel1Procedure:
			dRKeyIntraServiceDRKeyIntraLevel1Handler.ServeHTTP(w, r)
		case DRKeyIntraServiceDRKeyASHostProcedure:
			dRKeyIntraServiceDRKeyASHostHandler.ServeHTTP(w, r)
		case DRKeyIntraServiceDRKeyHostASProcedure:
			dRKeyIntraServiceDRKeyHostASHandler.ServeHTTP(w, r)
		case DRKeyIntraServiceDRKeyHostHostProcedure:
			dRKeyIntraServiceDRKeyHostHostHandler.ServeHTTP(w, r)
		case DRKeyIntraServiceDRKeySecretValueProcedure:
			dRKeyIntraServiceDRKeySecretValueHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedDRKeyIntraServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedDRKeyIntraServiceHandler struct{}

func (UnimplementedDRKeyIntraServiceHandler) DRKeyIntraLevel1(context.Context, *connect.Request[control_plane.DRKeyIntraLevel1Request]) (*connect.Response[control_plane.DRKeyIntraLevel1Response], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("proto.control_plane.v1.DRKeyIntraService.DRKeyIntraLevel1 is not implemented"))
}

func (UnimplementedDRKeyIntraServiceHandler) DRKeyASHost(context.Context, *connect.Request[control_plane.DRKeyASHostRequest]) (*connect.Response[control_plane.DRKeyASHostResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("proto.control_plane.v1.DRKeyIntraService.DRKeyASHost is not implemented"))
}

func (UnimplementedDRKeyIntraServiceHandler) DRKeyHostAS(context.Context, *connect.Request[control_plane.DRKeyHostASRequest]) (*connect.Response[control_plane.DRKeyHostASResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("proto.control_plane.v1.DRKeyIntraService.DRKeyHostAS is not implemented"))
}

func (UnimplementedDRKeyIntraServiceHandler) DRKeyHostHost(context.Context, *connect.Request[control_plane.DRKeyHostHostRequest]) (*connect.Response[control_plane.DRKeyHostHostResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("proto.control_plane.v1.DRKeyIntraService.DRKeyHostHost is not implemented"))
}

func (UnimplementedDRKeyIntraServiceHandler) DRKeySecretValue(context.Context, *connect.Request[control_plane.DRKeySecretValueRequest]) (*connect.Response[control_plane.DRKeySecretValueResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("proto.control_plane.v1.DRKeyIntraService.DRKeySecretValue is not implemented"))
}
