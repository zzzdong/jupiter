package grpc_test

import (
	"context"
	"testing"
	"time"

	grpc_client "github.com/douyu/jupiter/pkg/client/grpc"
	"github.com/douyu/jupiter/tests/proto/testproto"
	mock_testproto "github.com/douyu/jupiter/tests/proto/testproto/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestBase test direct dial with New()
func TestDirectGrpc(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	var StatusFoo = status.Errorf(codes.DataLoss, "invalid request")
	mockClient := mock_testproto.NewMockGreeterClient(controller)
	t.Run("normal request", func(t *testing.T) {
		mockClient.EXPECT().
			SayHello(context.Background(), &testproto.HelloRequest{Name: "hello"}).
			Return(&testproto.HelloReply{Message: "fantacy"}, error(nil))
		reply, err := mockClient.SayHello(context.Background(), &testproto.HelloRequest{Name: "hello"})
		assert.Nil(t, err)
		assert.Equal(t, "fantacy", reply.Message)
	})
	t.Run("not normal request", func(t *testing.T) {
		mockClient.EXPECT().
			SayHello(context.Background(), &testproto.HelloRequest{Name: "panic"}).
			Return(nil, StatusFoo)
		reply, err := mockClient.SayHello(context.Background(), &testproto.HelloRequest{Name: "panic"})
		assert.Nil(t, reply)
		assert.Equal(t, StatusFoo.Error(), err.Error())
	})
	t.Run("slow response", func(t *testing.T) {
		mockClient.EXPECT().
			SayHello(context.Background(), &testproto.HelloRequest{Name: "slow"}).
			DoAndReturn(func(context.Context, *testproto.HelloRequest) (*testproto.HelloReply, error) {
				time.Sleep(time.Millisecond * 600)
				return &testproto.HelloReply{Message: "fantacy"}, error(nil)
			})
		reply, err := mockClient.SayHello(context.Background(), &testproto.HelloRequest{Name: "slow"})
		assert.Nil(t, err)
		assert.Equal(t, "fantacy", reply.Message)

		// todo(gorexlv): 检测延时
	})
}

func TestGRPCBalance(t *testing.T) {

}

func TestConfigBuild(t *testing.T) {
	t.Run("test no address block, and panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("recover %+v", r)
			}
		}()
		cfg := grpc_client.DefaultConfig()
		cfg.OnDialError = "panic"
		cfg.Build()
	})
	t.Run("test no address and no block", func(t *testing.T) {
		cfg := grpc_client.DefaultConfig()
		cfg.OnDialError = "panic"
		cfg.Block = false
		conn := cfg.Build()
		assert.Equal(t, "IDLE", conn.GetState().String())
	})
}
