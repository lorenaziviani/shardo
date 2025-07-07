package grpcserver

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"shardo/pkg/cache"
)

type CacheServiceServer interface {
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Set(context.Context, *SetRequest) (*SetResponse, error)
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	Metrics(context.Context, *MetricsRequest) (*MetricsResponse, error)
}

type server struct {
	cache *cache.Cache
}

type GetRequest struct {
	Key string
}
type GetResponse struct {
	Value []byte
	Found bool
}
type SetRequest struct {
	Key   string
	Value []byte
	Ttl   int64 // seconds
}
type SetResponse struct{}
type DeleteRequest struct {
	Key string
}
type DeleteResponse struct{}
type MetricsRequest struct{}
type MetricsResponse struct {
	Hits   int32
	Misses int32
	Size   int32
}

func (s *server) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	val, ok := s.cache.Get(req.Key)
	return &GetResponse{Value: val, Found: ok}, nil
}
func (s *server) Set(ctx context.Context, req *SetRequest) (*SetResponse, error) {
	s.cache.Set(req.Key, req.Value, time.Duration(req.Ttl)*time.Second)
	return &SetResponse{}, nil
}
func (s *server) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	s.cache.Delete(req.Key)
	return &DeleteResponse{}, nil
}
func (s *server) Metrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error) {
	hits, misses, size := s.cache.Metrics()
	return &MetricsResponse{Hits: int32(hits), Misses: int32(misses), Size: int32(size)}, nil
}

func StartGRPCServer(port string, cacheCap int) {
	s := grpc.NewServer()

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("gRPC cache node listening on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
