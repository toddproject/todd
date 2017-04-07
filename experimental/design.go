package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = API("adder", func() {
	Title("The adder API")
	Description("A teaser for goa")
	Host("localhost:8080")
	Scheme("http")
})

var _ = Resource("operands", func() {
	Action("add", func() {
		Routing(GET("add/:left/:right"))
		Description("add returns the sum of the left and right parameters in the response body")
		Params(func() {
			Param("left", Integer, "Left operand")
			Param("right", Integer, "Right operand")
		})
		Response(OK, "text/plain")
	})

})

// http.HandleFunc("/v1/agent", tapi.Agent)
// http.HandleFunc("/v1/groups", tapi.Groups)
// http.HandleFunc("/v1/object/list", tapi.ListObjects)
// http.HandleFunc("/v1/object/group", tapi.ListObjects)
// http.HandleFunc("/v1/object/testrun", tapi.ListObjects)
// http.HandleFunc("/v1/object/create", tapi.CreateObject)
// http.HandleFunc("/v1/object/delete", tapi.DeleteObject)
// http.HandleFunc("/v1/testrun/run", tapi.Run)
// http.HandleFunc("/v1/testdata", tapi.TestData)
