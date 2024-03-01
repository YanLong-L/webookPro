package main

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	intrv1 "webookpro/api/proto/gen/intr/v1"
)

func TestGRPCClient(t *testing.T) {
	cc, err := grpc.Dial("localhost:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := intrv1.NewInteractiveServiceClient(cc)
	resp, err := client.Get(context.Background(), &intrv1.GetRequest{
		Biz:   "test",
		BizId: 2,
		Uid:   345,
	})
	require.NoError(t, err)
	t.Log(resp.Intr)
}

func TestGRPCDoubleWrite(t *testing.T) {
	// 写个 for 循环来模拟
	cc, err := grpc.Dial("localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := intrv1.NewInteractiveServiceClient(cc)
	_, err = client.IncrReadCnt(context.Background(), &intrv1.IncrReadCntRequest{
		Biz:   "test",
		BizId: 2,
	})
	require.NoError(t, err)
}
