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

package im

import (
    "fmt"

    "github.com/go-roc/roc/_auxiliary/example/tutorials/proto/pbim"
    "github.com/go-roc/roc/parcel/context"
)

type Im struct {
	H *Hub
	p *point
}

func (i *Im) Connect(c *context.Context, req *pbim.ConnectReq) (rsp *pbim.ConnectRsp, err error) {
	i.p = &point{userName: req.UserName, message: make(chan *pbim.SendMessageRsp)}
	i.H.addClient(i.p)
	return &pbim.ConnectRsp{IsConnect: true}, nil
}

func (i *Im) Count(c *context.Context, req *pbim.CountReq) (rsp *pbim.CountRsp, err error) {
	return &pbim.CountRsp{
		Count: i.H.count(),
	}, nil
}

func (i *Im) SendMessage(c *context.Context, req chan *pbim.SendMessageReq, errIn chan error) (chan *pbim.SendMessageRsp, chan error) {
	var rsp = make(chan *pbim.SendMessageRsp)
	var errs = make(chan error)

	go func() {
		for data := range i.p.message {
			rsp <- data
		}
		close(rsp)
	}()

	go func() {
	QUIT:
		for {
			select {
			case data, ok := <-req:
				if !ok {
					break QUIT
				}

				i.H.broadCast <- data

			case e := <-errIn:
				if e != nil {
					fmt.Println("----------close------")
					errs <- e
					break QUIT
				}
			}
		}

		close(errs)
		i.H.removeClient(i.p)
	}()

	return rsp, errs
}
