/*
Copyright 2017 The Kubernetes Authors.

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

package annotations

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-gce/pkg/flags"
)

func TestService(t *testing.T) {
	for _, tc := range []struct {
		svc             *v1.Service
		appProtocolsErr bool
		appProtocols    map[string]AppProtocol
		neg             bool
		http2           bool
	}{
		{
			svc:          &v1.Service{},
			appProtocols: map[string]AppProtocol{},
		},
		{
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceApplicationProtocolKey: `{"80": "HTTP", "443": "HTTPS"}`,
					},
				},
			},
			appProtocols: map[string]AppProtocol{"80": "HTTP", "443": "HTTPS"},
		},
		{
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceApplicationProtocolKey: `{"443": "HTTP2"}`,
					},
				},
			},
			appProtocols:    map[string]AppProtocol{"443": "HTTP2"},
			appProtocolsErr: true, // Without the http2 flag enabled, expect error
		},
		{
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceApplicationProtocolKey: `{"443": "HTTP2"}`,
					},
				},
			},
			appProtocols: map[string]AppProtocol{"443": "HTTP2"},
			http2:        true,
		},
		{
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						NetworkEndpointGroupAlphaAnnotation: "true",
					},
				},
			},
			appProtocols: map[string]AppProtocol{},
			neg:          true,
		},
		{
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceApplicationProtocolKey: `invalid`,
					},
				},
			},
			appProtocolsErr: true,
		},
		{
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceApplicationProtocolKey: `{"SSH": "22"}`,
					},
				},
			},
			appProtocolsErr: true,
		},
	} {
		flags.F.Features.Http2 = tc.http2
		svc := FromService(tc.svc)
		ap, err := svc.ApplicationProtocols()
		if tc.appProtocolsErr {
			if err == nil {
				t.Errorf("for service %+v; svc.ApplicationProtocols() = _, %v; want _, error", tc.svc, err)
			}
			continue
		}
		if err != nil || !reflect.DeepEqual(ap, tc.appProtocols) {
			t.Errorf("for service %+v; svc.ApplicationProtocols() = %v, %v; want %v, nil", tc.svc, ap, err, tc.appProtocols)
		}
		if b := svc.NEGEnabled(); b != tc.neg {
			t.Errorf("for service %+v; svc.NEGEnabled() = %v; want %v", tc.svc, b, tc.neg)
		}
	}
}
