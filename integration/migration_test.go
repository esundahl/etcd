// Copyright 2015 CoreOS, Inc.
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

package integration

import (
	"os/exec"
	"testing"
	"time"

	"github.com/coreos/etcd/client"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
)

func TestUpgradeMember(t *testing.T) {
	defer afterTest(t)
	m := mustNewMember(t, "integration046")
	cmd := exec.Command("cp", "-r", "testdata/integration046_data/conf", "testdata/integration046_data/log", "testdata/integration046_data/snapshot", m.DataDir)
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Launch(); err != nil {
		t.Fatal(err)
	}
	defer m.Terminate(t)
	// wait member recovered
	cc := mustNewHTTPClient(t, []string{m.URL()})
	ma := client.NewMembersAPI(cc)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		ms, err := ma.List(ctx)
		cancel()
		wmemb := client.Member{
			// name and PeerURLs are inherited from old data dir
			Name:       "integration046",
			PeerURLs:   []string{"http://127.0.0.1:59892"},
			ClientURLs: []string{m.URL()},
		}
		if err == nil && isMembersEqual(ms, []client.Member{wmemb}) {
			break
		}
		time.Sleep(tickDuration)
	}

	// check the data has been migrated
	ka := client.NewKeysAPI(cc)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := ka.Get(ctx, "qux", nil)
	cancel()
	if err != nil {
		t.Fatalf("watch error: %v", err)
	}
	if resp.Node.Value != "quux" {
		t.Errorf("value(qux) = %s, want quux", resp.Node.Value)
	}

	clusterMustProgress(t, []*member{m})
}
