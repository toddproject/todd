/*
   ToDD agent responses

   These are asynchronous responses sent back to the server, usually as a response to a task.
   Not all tasks result in an agent sending a response to the server. Responses are for highly sensitive operations like test distribution and execution.

    Copyright 2016 Matt Oswalt. Use or modification of this
    source code is governed by the license provided here:
    https://github.com/toddproject/todd/blob/master/LICENSE
*/

package responses

// Response is an interface to define Response behavior
type Response interface{}

// BaseResponse is a struct that is intended to be embedded by specific response structs. Both of these in conjunction
// are used primarily to house the JSON message for passing responses over the comms package (i.e. message queue), but may also contain important
// dependencies of the response, such as an HTTP handler.
type BaseResponse struct {
	AgentUuid string `json:"agentuuid"`
	Type      string `json:"type"`
}
