// Copyright (c) 2021 roc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//

package hello

import (
    "net/http"
    "time"

    "github.com/coreos/etcd/clientv3"

    "github.com/go-roc/roc"
    "github.com/go-roc/roc/_auxiliary/example/tutorials/proto/pbhello"
    "github.com/go-roc/roc/parcel/context"
    "github.com/go-roc/roc/rlog"
)

type Hello struct {
    opt    roc.InvokeOptions
    client pbhello.HelloWorldClient
}

// NewHello new Hello and initialize it for rpc client
// opt is configurable when request.
func NewHello() *Hello {
    return &Hello{
        client: pbhello.NewHelloWorldClient(
            roc.NewService(
                roc.EtcdConfig(
                    &clientv3.Config{
                        Endpoints: []string{"127.0.0.1:2379"},
                    },
                ),
            ),
        ),
        //opt: roc.WithAddress("srv.hello","169.254.241.13:16856"),
        opt: roc.WithName("srv.hello"),
    }
}

func (h *Hello) SayHandler(w http.ResponseWriter, r *http.Request) {
    _ = r.ParseForm()

    now := time.Now()
    rsp, err := h.client.Say(
        context.Background(),
        &pbhello.SayReq{Inc: 1},
        h.opt,
    )
    if err != nil {
        w.Write([]byte(err.Error()))
        return
    }

    rlog.Infof("FROM hello server: %v |latency=%v ms ", rsp.Inc, time.Since(now).Milliseconds())

    w.Write([]byte("succuess"))
}
