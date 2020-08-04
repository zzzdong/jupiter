package etcdv3

import (
	"context"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	mock_clientv3 "github.com/douyu/jupiter/pkg/client/etcdv3/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_GetAndPut(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockKV := mock_clientv3.NewMockKV(controller)
	mockLease := mock_clientv3.NewMockLease(controller)

	client := &Client{
		Client: &clientv3.Client{
			KV:    mockKV,
			Lease: mockLease,
		},
	}

	t.Run("get with key", func(t *testing.T) {
		mockKV.EXPECT().Get(context.Background(), "/test/key").Return("value1")

		kv, err := client.GetKeyValue(context.Background(), "/test/key")
		assert.Nil(t, err)
		assert.Equal(t, "value1", string(kv.Value))
	})

	t.Run("get with prefix", func(t *testing.T) {

	})

	t.Run("get with lease", func(t *testing.T) {

	})

	t.Run("put with lease", func(t *testing.T) {

	})
}

func Test_MutexLock(t *testing.T) {
	config := DefaultConfig()
	config.Endpoints = []string{"127.0.0.1:2379"}
	config.TTL = 10
	etcdCli := newClient(config)

	etcdMutex1, err := etcdCli.NewMutex("/test/lock",
		concurrency.WithTTL(int(config.TTL)))
	assert.Nil(t, err)

	err = etcdMutex1.Lock(time.Second * 1)
	assert.Nil(t, err)
	defer etcdMutex1.Unlock()

	// Grab the lock
	etcdMutex, err := etcdCli.NewMutex("/test/lock",
		concurrency.WithTTL(int(config.TTL)))
	assert.Nil(t, err)
	defer etcdMutex.Unlock()

	err = etcdMutex.Lock(time.Second * 1)
	assert.NotNil(t, err)
}
