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

package namespace

type Service = string

const (
	DefaultVersion      Service = "v1.0.0"
	DefaultSchema               = "goroc"
	DefaultConfigSchema         = "configroc"
)

type Header = string

const (
	DefaultHeaderTrace   Header = "X-Idempotency-Key"
	DefaultHeaderVersion        = "X-Api-Version"
	DefaultHeaderToken          = "X-Api-Token"
	DefaultHeaderAddress        = "X-Api-Address"
)

type Schema = string

type Scope = string

type RequestChannel = string
