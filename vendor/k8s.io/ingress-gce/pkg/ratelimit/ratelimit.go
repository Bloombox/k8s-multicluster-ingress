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
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud/meta"
)

// GCERateLimiter implements cloud.RateLimiter
type GCERateLimiter struct {
	// Map a RateLimitKey to its rate limiter implementation.
	rateLimitImpls map[cloud.RateLimitKey]flowcontrol.RateLimiter
}

// NewGCERateLimiter parses the list of rate limiting specs passed in and
// returns a properly configured cloud.RateLimiter implementation.
// Expected format of specs: {"[version].[service].[operation],[type],[param1],[param2],..", "..."}
func NewGCERateLimiter(specs []string) (*GCERateLimiter, error) {
	rateLimitImpls := make(map[cloud.RateLimitKey]flowcontrol.RateLimiter)
	// Within each specification, split on comma to get the operation,
	// rate limiter type, and extra parameters.
	for _, spec := range specs {
		params := strings.Split(spec, ",")
		if len(params) < 2 {
			return nil, fmt.Errorf("Must at least specify operation and rate limiter type.")
		}
		// params[0] should consist of the operation to rate limit.
		key, err := constructRateLimitKey(params[0])
		if err != nil {
			return nil, err
		}
		// params[1:] should consist of the rate limiter type and extra params.
		impl, err := constructRateLimitImpl(params[1:])
		if err != nil {
			return nil, err
		}
		rateLimitImpls[key] = impl
		glog.Infof("Configured rate limiting for: %v", key)
	}
	if len(rateLimitImpls) == 0 {
		return nil, nil
	}
	return &GCERateLimiter{rateLimitImpls}, nil
}

// Implementation of cloud.RateLimiter
func (l *GCERateLimiter) Accept(ctx context.Context, key *cloud.RateLimitKey) error {
	ch := make(chan struct{})
	go func() {
		// Call flowcontrol.RateLimiter implementation.
		impl := l.rateLimitImpl(key)
		if impl != nil {
			impl.Accept()
		}
		close(ch)
	}()
	select {
	case <-ch:
		break
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// rateLimitImpl returns the flowcontrol.RateLimiter implementation
// associated with the passed in key.
func (l *GCERateLimiter) rateLimitImpl(key *cloud.RateLimitKey) flowcontrol.RateLimiter {
	// Since the passed in key will have the ProjectID field filled in, we need to
	// create a copy which does not, so that retreiving the rate limiter implementation
	// through the map works as expected.
	keyCopy := cloud.RateLimitKey{
		ProjectID: "",
		Operation: key.Operation,
		Version:   key.Version,
		Service:   key.Service,
	}
	return l.rateLimitImpls[keyCopy]
}

// Expected format of param is [version].[service].[operation]
func constructRateLimitKey(param string) (cloud.RateLimitKey, error) {
	var retVal cloud.RateLimitKey
	params := strings.Split(param, ".")
	if len(params) != 3 {
		return retVal, fmt.Errorf("Must specify rate limit in [version].[service].[operation] format: %v", param)
	}
	// TODO(rramkumar): Add another layer of validation here?
	version := meta.Version(params[0])
	service := params[1]
	operation := params[2]
	retVal = cloud.RateLimitKey{
		ProjectID: "",
		Operation: operation,
		Version:   version,
		Service:   service,
	}
	return retVal, nil
}

// constructRateLimitImpl parses the slice and returns a flowcontrol.RateLimiter
// Expected format is [type],[param1],[param2],...
func constructRateLimitImpl(params []string) (flowcontrol.RateLimiter, error) {
	// For now, only the "qps" type is supported.
	rlType := params[0]
	implArgs := params[1:]
	if rlType == "qps" {
		if len(implArgs) != 2 {
			return nil, fmt.Errorf("Invalid number of args for rate limiter type %v. Expected %d, Got %v", rlType, 2, len(implArgs))
		}
		qps, err := strconv.ParseFloat(implArgs[0], 32)
		if err != nil || qps <= 0 {
			return nil, fmt.Errorf("Invalid argument for rate limiter type %v. Either %v is not a float or not greater than 0.", rlType, implArgs[0])
		}
		burst, err := strconv.Atoi(implArgs[1])
		if err != nil {
			return nil, fmt.Errorf("Invalid argument for rate limiter type %v. Expected %v to be a int.", rlType, implArgs[1])
		}
		return flowcontrol.NewTokenBucketRateLimiter(float32(qps), burst), nil
	}
	return nil, fmt.Errorf("Invalid rate limiter type provided: %v", rlType)
}
