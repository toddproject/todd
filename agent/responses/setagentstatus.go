/*
   ToDD response - set agent status

    Copyright 2016 Matt Oswalt. Use or modification of this
    source code is governed by the license provided here:
    https://github.com/toddproject/todd/blob/master/LICENSE
*/

package responses

// SetAgentStatusResponse defines this particular response.
type SetAgentStatusResponse struct {
	BaseResponse
	TestUuid string `json:"TestUuid"`
	Status   string `json:"status"`
}
