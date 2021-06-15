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

package client

import (
	"errors"
	"sync"
	"sync/atomic"

	"roc/internal/endpoint"
	"roc/internal/namespace"
	"roc/internal/registry"
	"roc/internal/transport"
	"roc/internal/x"
	"roc/rlog"
)

var (
	ErrorNoneServer   = errors.New("server is none to use")
	ErrorNoSuchServer = errors.New("no such server")
)

type Strategy interface {

	//Next Round-robin scheduling
	Next(scope string) (next *Conn, err error)

	//Straight direct call
	Straight(scope, address string) (next *Conn, err error)

	//Close Strategy
	Close()
}

var _ Strategy = &strategy{}

type strategy struct {
	sync.Mutex

	//per service & multiple conn
	connPerService map[string]*pod

	//discover registry
	registry registry.Registry

	//transport client
	client transport.Client

	//registry watch callback action
	action chan *registry.Action

	//close strategy signal
	close chan struct{}
}

// create a strategy
func newStrategy(registry registry.Registry, client transport.Client) Strategy {
	s := &strategy{
		connPerService: make(map[string]*pod),
		registry:       registry,
		client:         client,
		close:          make(chan struct{}),
	}

	//receive registry watch notify
	s.action = s.registry.Watch()

	//Synchronize all existing services
	s.lazySync()

	//handler registry notify
	go s.notify()

	return s
}

//get a pod,if is nil ,create a new pod
func (s *strategy) getOrSet(scope string) (*pod, error) {
	p, ok := s.connPerService[scope]
	if !ok || p.count == 0 {
		s.Lock()
		defer s.Unlock()

		e, err := s.registry.Next(scope)
		if err != nil {
			return nil, err
		}
		err = s.sync(e)
		if err != nil {
			return nil, err
		}
		return s.connPerService[e.Scope], nil
	}
	//pod must available
	return p, nil
}

// Next Round-robin next
func (s *strategy) Next(scope string) (next *Conn, err error) {
	p, err := s.getOrSet(scope)
	if err != nil {
		return nil, err
	}

	var conn *Conn
	for i := 0; i < p.count; i++ {
		conn = p.clients[(int(atomic.AddUint32(&p.index, 1))-1)%p.count]
		if conn.isWorking() {
			break
		}
	}

	if conn == nil || !conn.isWorking() {
		return nil, ErrorNoneServer
	}

	return conn, nil
}

// Straight direct invoke
func (s *strategy) Straight(scope, address string) (next *Conn, err error) {
	p, err := s.getOrSet(scope)
	if err != nil {
		return nil, err
	}

	conn, ok := p.clientsMap[address]
	if !ok || !conn.isWorking() {
		return nil, ErrorNoSuchServer
	}

	return conn, nil
}

//Synchronize all existing services
func (s *strategy) lazySync() {
	s.Lock()
	defer s.Unlock()

	es, err := s.registry.List()
	if err != nil {
		rlog.Debug(err)
		return
	}

	for _, e := range es {
		s.sync(e)
	}
}

//Synchronize one services
func (s *strategy) sync(e *endpoint.Endpoint) error {
	p, ok := s.connPerService[e.Scope]
	if !ok {
		p = newPod()
	}

	err := p.Add(e, s.client)
	if err != nil {
		rlog.Error(err)
		return err
	}

	s.connPerService[e.Scope] = p

	return nil
}

//receive a registry notify callback
func (s *strategy) notify() {

QUIT:
	for {
		select {
		case act := <-s.action:
			s.watch(act)
			rlog.Debug("watch endpoint.Endpoint was changed", x.MustMarshalString(act))
		case <-s.close:
			break QUIT
		}
	}
}

//watch a registry notify
func (s *strategy) watch(act *registry.Action) {
	s.Lock()
	defer s.Unlock()

	switch act.Act {
	case namespace.WatcherCreate:
		_ = s.sync(act.E)
	case namespace.WatcherDelete:
		p, ok := s.connPerService[act.E.Scope]
		if ok {
			p.Del(act.E.Address)
		}
	}
}

// Close strategy close
func (s *strategy) Close() {
	for _, p := range s.connPerService {
		for _, client := range p.clientsMap {
			client.Close()
		}
	}

	s.close <- struct{}{}
}
