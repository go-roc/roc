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

package config

import (
	"github.com/go-roc/roc/internal/etcd"
	"github.com/go-roc/roc/internal/namespace"
	"github.com/go-roc/roc/internal/x"
)

type Option struct {

	//etcd
	e *etcd.Etcd

	//disableDynamic switch
	disableDynamic bool

	//config schema on etcd
	//schema usually is public config dir
	schema string

	//private config on etcd
	private string

	//public config on etcd
	public string

	//public prefix
	prefix string

	//config version
	version string

	//backup file path
	backupPath string

	//localFile switch
	//if true,will load local config.json to
	localFile bool

	//f will be run after config already setup
	f []func() error
}

type Options func(option *Option)

//Chaos is config already setup and do Chaos functions
func Chaos(f ...func() error) Options {
	return func(option *Option) {
		option.f = f
	}
}

func Private(private string) Options {
	return func(option *Option) {
		option.private = private
	}
}

func Public(public string) Options {
	return func(option *Option) {
		option.public = public
	}
}

func LocalFile() Options {
	return func(option *Option) {
		option.localFile = true
	}
}

func DisableDynamic() Options {
	return func(option *Option) {
		option.disableDynamic = true
	}
}

func Schema(schema string) Options {
	return func(option *Option) {
		option.schema = schema
	}
}

func Version(version string) Options {
	return func(option *Option) {
		option.version = version
	}
}

func Backup(path string) Options {
	return func(option *Option) {
		option.backupPath = path
	}
}

func Prefix(prefix string) Options {
	return func(option *Option) {
		option.private = prefix
	}
}

func newOpts(opts ...Options) Option {
	opt := Option{}

	for i := range opts {
		opts[i](&opt)
	}

	if opt.schema == "" {
		opt.schema = namespace.DefaultConfigSchema
	}

	if opt.version == "" {
		opt.version = namespace.DefaultVersion
	}

	opt.schema += "/" + opt.version

	if opt.backupPath == "" {
		opt.backupPath = "./config.json"
	}

	if opt.private == "" {
		opt.private = x.GetProjectName()
	}

	opt.private = opt.schema + "/" + opt.private + "/"

	if opt.public == "" {
		opt.public = "public"
	}

	opt.public = opt.schema + "/" + opt.public + "/"

	opt.e = etcd.DefaultEtcd

	if opt.prefix == "" {
		opt.prefix = "roc."
	}

	return opt
}
