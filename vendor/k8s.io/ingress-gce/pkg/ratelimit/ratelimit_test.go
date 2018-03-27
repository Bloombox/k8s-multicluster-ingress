/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ratelimit

import (
	"testing"
)

func TestConfigureGCERateLimiting(t *testing.T) {
	validTestCases := [][]string{
		[]string{"ga.Addresses.Get,qps,1.5,5"},
		[]string{"ga.Addresses.List,qps,2,10"},
		[]string{"ga.Addresses.Get,qps,1.5,5", "ga.Firewalls.Get,qps,1.5,5"},
		[]string{"ga.Operations.Get,qps,10,100"},
	}
	invalidTestCases := [][]string{
		[]string{"gaAddresses.Get,qps,1.5,5"},
		[]string{"gaAddresses.Get,qps,0,5"},
		[]string{"gaAddresses.Get,qps,-1,5"},
		[]string{"ga.Addresses.Get,qps,1.5.5"},
		[]string{"gaAddresses.Get,qps,1.5,5.5"},
		[]string{"gaAddressesGet,qps,1.5,5.5"},
		[]string{"gaAddressesGet,qps,1.5"},
		[]string{"ga.Addresses.Get,foo,1.5,5"},
		[]string{"ga.Addresses.Get,1.5,5"},
		[]string{"ga.Addresses.Get,qps,1.5,5", "gaFirewalls.Get,qps,1.5,5"},
	}

	for _, testCase := range validTestCases {
		_, err := NewGCERateLimiter(testCase)
		if err != nil {
			t.Errorf("Did not expect an error for test case: %v", testCase)
		}
	}

	for _, testCase := range invalidTestCases {
		_, err := NewGCERateLimiter(testCase)
		if err == nil {
			t.Errorf("Expected an error for test case: %v", testCase)
		}
	}
}
