package api

import (
	"context"
	"fmt"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"

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
