/*
   Grouping tests

   Copyright 2017 Matt Oswalt. Use or modification of this
   source code is governed by the license provided here:
   https://github.com/toddproject/todd/blob/master/LICENSE
*/

package grouping

import (
	"testing"
)

func TestIsInGroup(t *testing.T) {
	tests := []struct {
		label           string
		matchStatements []map[string]string
		factmap         map[string][]string
		want            bool
	}{
		{
			label: "hostname not in group",
			matchStatements: []map[string]string{
				{"hostname": "toddtestagent[1-6]"},
			},
			factmap: map[string][]string{
				"Addresses": {"127.0.0.1", "::1"},
				"Hostname":  {"toddtestagent7"},
			},
			want: false,
		},
		{
			label: "hostname in group",
			matchStatements: []map[string]string{
				{"hostname": "toddtestagent[1-6]"},
			},
			factmap: map[string][]string{
				"Addresses": {"127.0.0.1", "::1"},
				"Hostname":  {"toddtestagent6"},
			},
			want: true,
		},
		{
			label: "multiple hostnames",
			matchStatements: []map[string]string{
				{"hostname": "toddtestagent1"},
				{"hostname": "toddtestagent2"},
				{"hostname": "toddtestagent3"},
				{"hostname": "toddtestagent4"},
				{"hostname": "toddtestagent5"},
				{"hostname": "toddtestagent6"},
			},
			factmap: map[string][]string{
				"Addresses": {"127.0.0.1", "::1"},
				"Hostname":  {"toddtestagent6"},
			},
			want: true,
		},
		{
			label: "subnet not in group",
			matchStatements: []map[string]string{
				{"within_subnet": "192.168.0.0/24"},
			},
			factmap: map[string][]string{
				"Addresses": {"127.0.0.1", "::1"},
				"Hostname":  {"todddev"},
			},
			want: false,
		},
		{
			label: "subnet in group",
			matchStatements: []map[string]string{
				{"within_subnet": "192.168.0.0/24"},
			},
			factmap: map[string][]string{
				"Addresses": {"192.168.0.1", "::1"},
				"Hostname":  {"todddev"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			got := isInGroup(tt.matchStatements, tt.factmap)
			if got != tt.want {
				t.Errorf("expected isInGroup(%+v, %+v) to be %t, but it was %t", tt.matchStatements, tt.factmap, tt.want, got)
			}
		})
	}
}
