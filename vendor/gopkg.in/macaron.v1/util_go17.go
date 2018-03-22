// +build !go1.8

// Copyright 2017 The Macaron Authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package macaron

import "net/url"

// PathUnescape unescapes a path. Ideally, this function would use
// url.PathUnescape(..), but the function was not introduced until go1.8.
func PathUnescape(s string) (string, error) {
	return url.QueryUnescape(s)
}
