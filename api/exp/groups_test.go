package api

import (
	"context"
	"fmt"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc"

	pb "github.com/toddproject/todd/api/exp/generated"
)

// What to test
/*

- All malformed definitions are rejected, correctly formed are accepted
- Single creation, single "get" works.
- Multiple creation, "list" matches the count created (try with varying lengths)

*/

// TestGroupsValidation ensures that the protobuf validation rules are effective at blocking malformed
// group definitions at the API level.
func TestGroupsValidation(t *testing.T) {

	// Start client
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:50099"), grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	client := pb.NewGroupsClient(conn)

	// Test parameters
	type apiTestResult struct {
		inputGroup     *pb.Group
		expectedResult string
	}
	groupTestTable := []apiTestResult{
		{
			// Should fail because the name is too short
			inputGroup: &pb.Group{
				Name: "fo",
			},
			expectedResult: "invalid Group.Name: value length must be at least 3 runes",
		},
		{
			// Should fail because there are no matches
			inputGroup: &pb.Group{
				Name: "foobar",
			},
			expectedResult: "invalid Group.Matches: value must contain at least 1 item(s)",
		},
		{
			// Should fail because match type isn't "hostname" or "subnet"
			inputGroup: &pb.Group{
				Name: "foobar",
				Matches: []*pb.Group_Match{
					{
						Type:      "foobar",
						Statement: "barfoo",
					},
				},
			},
			expectedResult: "invalid Group_Match.Type: value must be in list [hostname subnet]",
		},
		{
			// Should fail because the statement field is missing
			inputGroup: &pb.Group{
				Name: "foobar",
				Matches: []*pb.Group_Match{
					{
						Type: "hostname",
					},
				},
			},
			expectedResult: "invalid Group_Match.Statement: value length must be at least 3 runes",
		},

		// TODO(mierdin) Add validation tests for specific match rule semantics

		{
			// Should succeed; valid group
			inputGroup: &pb.Group{
				Name: "foobar",
				Matches: []*pb.Group_Match{
					{
						Type:      "hostname",
						Statement: "foobar",
					},
				},
			},
			expectedResult: "",
		},
	}

	for t := range groupTestTable {
		tr := groupTestTable[t]

		_, err = client.CreateGroup(context.Background(), tr.inputGroup)

		// Handle nil values
		if tr.expectedResult == "" && err != nil {
			log.Fatalf("Group validation %d test failed", t)
		} else if err == nil {
			continue
		}

		// Assuming not nil, validate error message
		if !strings.Contains(err.Error(), tr.expectedResult) {
			log.Fatalf("Group validation %d test failed", t)
		}
	}

}

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
