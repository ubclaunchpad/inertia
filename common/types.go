// Copyright © 2017 UBC Launch Pad team@ubclaunchpad.com
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

package common

const (
	// DefaultSecret used for some verification
	DefaultSecret = "inertia"

	// DaemonOkResp is the OK response upon successfully reaching daemon
	DaemonOkResp = "I'm a little Webhook, short and stout!"
)

// DaemonRequest is the configurable body of a request to the daemon.
type DaemonRequest struct {
	Stream    bool   `json:"stream"`
	Repo      string `json:"repo,omitempty"`
	Container string `json:"container,omitempty"`
}
