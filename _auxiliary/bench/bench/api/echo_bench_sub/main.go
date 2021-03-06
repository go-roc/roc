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

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/core/transport"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/mono"
)

var tp transport.ClientTransporter

func init() {
	tp = rsocket.TCPClient().SetHostAndPort("127.0.0.1", 7878).Build()
}

//windows
//cpu i9 9980hk 2.4hz
//mem 16gb
//使用多客户端
//-----------------------tps------------------- 178741
//-----------------------tps------------------- 189730
//-----------------------tps------------------- 203555
//-----------------------tps------------------- 204234
//-----------------------tps------------------- 204728
//-----------------------tps------------------- 197920
//-----------------------tps------------------- 207015

//使用单客户端
//-----------------------tps------------------- 344731
//-----------------------tps------------------- 346055
//-----------------------tps------------------- 343354
//-----------------------tps------------------- 342193
//-----------------------tps------------------- 348156
func main() {
	const (
		parallel = 100
		round    = 100
		amount   = 100000
	)

	var clients []rsocket.Client
	for i := 0; i < 32; i++ {
		client, err := createClient()
		if err != nil {
			return
		}
		clients = append(clients, client)
	}
	if len(clients) == 0 {
		return
	}

	defer func() {
		for _, c := range clients {
			_ = c.Close()
		}
	}()

	time.Sleep(time.Second * 1)

	var tps int64
	var errCount int64
	var next uint64
	for i := 0; i < 100; i++ {
		go func() {
			c := clients[(int(atomic.AddUint64(&next, 1))-1)%len(clients)]
			for j := 0; j < 100000; j++ {
				var s = make(chan struct{})
				var b string
				c.RequestResponse(payload.NewString(strconv.Itoa(j), "")).Subscribe(context.Background(), rx.OnComplete(func() {
					s <- struct{}{}
				}), rx.OnError(func(e error) {
					atomic.AddInt64(&errCount, 1)
				}), rx.OnNext(func(input payload.Payload) error {
					b = input.DataUTF8()
					return nil
				}))
				<-s
				if b != strconv.Itoa(j) {
					atomic.AddInt64(&errCount, 1)
				}
				atomic.AddInt64(&tps, 1)
			}
		}()
	}

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			fmt.Println("-----------------------tps------------------errorcount-----", atomic.LoadInt64(&tps), atomic.LoadInt64(&errCount))
			atomic.SwapInt64(&tps, 0)
		}
	}
}

func createClient() (rsocket.Client, error) {
	return rsocket.Connect().
		Scheduler(scheduler.Elastic(), nil).
		OnClose(func(err error) {
			log.Println("*** disconnected ***")
		}).
		SetupPayload(payload.NewString("你好", "世界")).
		Acceptor(func(ctx context.Context, socket rsocket.RSocket) rsocket.RSocket {
			return rsocket.NewAbstractSocket(
				rsocket.RequestResponse(func(p payload.Payload) mono.Mono {
					log.Println("receive request from server:", p)
					if bytes.Equal(p.Data(), []byte("ping")) {
						return mono.Just(payload.NewString("pong", "from client"))
					}
					return mono.Just(p)
				}),
			)
		}).
		Transport(tp).
		Start(context.Background())
}
