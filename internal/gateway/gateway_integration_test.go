package gateway

import (
	"context"
	"os"
	"testing"
	"time"

	"shardo/proto/cachepb"

	"google.golang.org/grpc"
)

func TestNodeGRPCSetGet(t *testing.T) {
	addr := os.Getenv("TEST_NODE_ADDR")
	if addr == "" {
		t.Skip("TEST_NODE_ADDR not set")
	}
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()
	client := cachepb.NewCacheServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.Set(ctx, &cachepb.SetRequest{Key: "foo", Value: []byte("bar"), Ttl: 60})
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	resp, err := client.Get(ctx, &cachepb.GetRequest{Key: "foo"})
	if err != nil || !resp.Found || string(resp.Value) != "bar" {
		t.Fatalf("Get failed: %v, found: %v, value: %s", err, resp.Found, resp.Value)
	}
}
