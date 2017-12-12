// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"testing"
)

func TestValidateListArgs(t *testing.T) {
	// It should return an error with extra args.
	options := listOptions{}
	if err := validateListArgs(&options, []string{"arg1"}); err == nil {
		t.Errorf("Expected error for non-empty args")
	}

	// It should return an error with missing project.
	options = listOptions{}
	if err := validateListArgs(&options, []string{}); err == nil {
		t.Errorf("Expected error for non-empty args")
	}

	// validateListArgs should succeed with a project and empty args.
	options = listOptions{
		GCPProject: "myGcpProject",
	}
	if err := validateListArgs(&options, []string{}); err != nil {
		t.Errorf("unexpected error from validateListArgs: %s", err)
	}
}
