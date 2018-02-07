package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

// These are the existing ToDD API functions that need to be covered
// http.HandleFunc("/v1/agent", tapi.Agent)
// http.HandleFunc("/v1/groups", tapi.Groups)
// http.HandleFunc("/v1/object/list", tapi.ListObjects)
// http.HandleFunc("/v1/object/group", tapi.ListObjects)
// http.HandleFunc("/v1/object/testrun", tapi.ListObjects)
// http.HandleFunc("/v1/object/create", tapi.CreateObject)
// http.HandleFunc("/v1/object/delete", tapi.DeleteObject)
// http.HandleFunc("/v1/testrun/run", tapi.Run)
// http.HandleFunc("/v1/testdata", tapi.TestData)

// Using https://github.com/goadesign/goa-cellar/tree/master/design as an example
var _ = API("todd", func() {
	Title("The ToDD API")
	Description("The ToDD API")
	Contact(func() {
		Name("Matt Oswalt")
		// Email("admin@goa.design")
		URL("https://github.com/toddproject")
	})
	License(func() {
		Name("Apache v2")
		URL("https://github.com/toddproject/todd/blob/master/LICENSE")
	})
	Docs(func() {

		// May want to also provide API docs?
		Description("todd docs")
		URL("https://todd.readthedocs.io")
	})
	Host("localhost:8081")
	Scheme("http")
	BasePath("/todd")

	Origin("http://swagger.goa.design", func() {
		Methods("GET", "POST", "PUT", "PATCH", "DELETE")
		MaxAge(600)
		Credentials()
	})

	ResponseTemplate(Created, func(pattern string) {
		Description("Resource created")
		Status(201)
		Headers(func() {
			Header("Location", String, "href to created resource", func() {
				Pattern(pattern)
			})
		})
	})
})
