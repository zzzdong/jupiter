// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcdv3

import (
	"context"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/douyu/jupiter/pkg/constant"
	"github.com/golang/mock/gomock"

	"github.com/douyu/jupiter/pkg/client/etcdv3"
	mock_clientv3 "github.com/douyu/jupiter/pkg/client/etcdv3/mock"
	"github.com/douyu/jupiter/pkg/server"
	"github.com/douyu/jupiter/pkg/xlog"
	"github.com/stretchr/testify/assert"
)

var (
	serviceA = &server.ServiceInfo{
		Name:       "service_1",
		AppID:      "",
		Scheme:     "grpc",
		Address:    "10.10.10.1:9091",
		Weight:     0,
		Enable:     true,
		Healthy:    true,
		Metadata:   map[string]string{},
		Region:     "default",
		Zone:       "default",
		Kind:       constant.ServiceProvider,
		Deployment: "default",
		Group:      "",
	}

	serviceB = &server.ServiceInfo{
		Name:       "service_1",
		AppID:      "",
		Scheme:     "grpc",
		Address:    "10.10.10.1:9092",
		Weight:     0,
		Enable:     true,
		Healthy:    true,
		Metadata:   map[string]string{},
		Region:     "default",
		Zone:       "default",
		Kind:       constant.ServiceProvider,
		Deployment: "default",
		Group:      "",
	}
)

func Test_etcdv3Registry(t *testing.T) {
	config := &Config{
		ReadTimeout: time.Second * 10,
		Prefix:      "jupiter",
		logger:      xlog.DefaultLogger,
	}
	controller := gomock.NewController(t)
	mockKV := mock_clientv3.NewMockKV(controller)
	config.WithETCDV3Client(&etcdv3.Client{
		Client: &clientv3.Client{
			KV: mockKV,
		},
	})
	registry := newETCDRegistry(config)

	t.Run("register service", func(t *testing.T) {
		realKey := registry.registerKey(serviceA)
		realVal := registry.registerValue(serviceA)
		// biz register
		mockKV.EXPECT().
			Put(gomock.Any(), realKey, realVal).
			DoAndReturn(func(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
				return &clientv3.PutResponse{}, nil
			})

		// metric register

		assert.Nil(t, registry.RegisterService(context.Background(), serviceA))
	})

	t.Run("list services", func(t *testing.T) {
		mockKV.EXPECT().
			Get(gomock.Any(), "/jupiter/service_1/providers/grpc://", gomock.Any()).
			DoAndReturn(func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return &clientv3.GetResponse{
					Header: &etcdserverpb.ResponseHeader{
						ClusterId: 0,
						MemberId:  0,
						Revision:  0,
						RaftTerm:  0,
					},
					Kvs: []*mvccpb.KeyValue{
						{Key: []byte(registry.registerKey(serviceA)), Value: []byte(registry.registerValue(serviceA))},
						{Key: []byte(registry.registerKey(serviceB)), Value: []byte(registry.registerValue(serviceB))},
					},
					More:  false,
					Count: 0,
				}, nil
			})
		services, err := registry.ListServices(context.Background(), "service_1", "grpc")
		assert.Nil(t, err)
		assert.Equal(t, 2, len(services))
		assert.Equal(t, "10.10.10.1:9091", services[0].Address)
		assert.Equal(t, "10.10.10.1:9092", services[1].Address)
	})

	// services, err := registry.ListServices(context.Background(), "service_1", "grpc")
	// assert.Nil(t, err)
	// assert.Equal(t, 1, len(services))
	// assert.Equal(t, "10.10.10.1:9091", services[0].Address)

	// go func() {
	// 	si := &server.ServiceInfo{
	// 		Name:       "service_1",
	// 		Scheme:     "grpc",
	// 		Address:    "10.10.10.1:9092",
	// 		Enable:     true,
	// 		Healthy:    true,
	// 		Metadata:   map[string]string{},
	// 		Region:     "default",
	// 		Zone:       "default",
	// 		Deployment: "default",
	// 	}
	// 	time.Sleep(time.Second)
	// 	assert.Nil(t, registry.RegisterService(context.Background(), si))
	// 	assert.Nil(t, registry.UnregisterService(context.Background(), si))
	// }()

	// ctx, cancel := context.WithCancel(context.Background())
	// go func() {
	// 	endpoints, err := registry.WatchServices(ctx, "service_1", "grpc")
	// 	assert.Nil(t, err)
	// 	for msg := range endpoints {
	// 		t.Logf("watch service: %+v\n", msg)
	// 		// 	assert.Equal(t, "10.10.10.2:9092", msg)
	// 	}
	// }()

	// time.Sleep(time.Second * 3)
	// cancel()
	// _ = registry.Close()
	// time.Sleep(time.Second * 1)
}

// func Test_etcdv3registry_UpdateAddressList(t *testing.T) {
// 	etcdConfig := etcdv3.DefaultConfig()
// 	etcdConfig.Endpoints = []string{"127.0.0.1:2379"}
// 	reg := newETCDRegistry(&Config{
// 		// Config:      etcdConfig,
// 		ReadTimeout: time.Second * 10,
// 		Prefix:      "jupiter",
// 		logger:      xlog.DefaultLogger,
// 	})

// 	var routeConfig = registry.RouteConfig{
// 		ID:         "1",
// 		Scheme:     "grpc",
// 		Host:       "",
// 		Deployment: "openapi",
// 		URI:        "/hello",
// 		Upstream: registry.Upstream{
// 			Nodes: map[string]int{
// 				"10.10.10.1:9091": 1,
// 				"10.10.10.1:9092": 10,
// 			},
// 		},
// 	}
// 	_, err := reg.client.Put(context.Background(), "/jupiter/service_1/configurators/grpc:///routes/1", routeConfig.String())
// 	assert.Nil(t, err)
// 	ctx, cancel := context.WithCancel(context.Background())
// 	go func() {
// 		services, err := reg.WatchServices(ctx, "service_1", "grpc")
// 		assert.Nil(t, err)
// 		fmt.Printf("len(services) = %+v\n", len(services))
// 		for service := range services {
// 			fmt.Printf("service = %+v\n", service)
// 		}
// 	}()
// 	time.Sleep(time.Second * 3)
// 	cancel()
// 	_ = reg.Close()
// 	time.Sleep(time.Second * 1)
// }
