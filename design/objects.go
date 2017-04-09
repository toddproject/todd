package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = Resource("object", func() {

	DefaultMedia(Object)
	BasePath("/object")

	Action("list", func() {
		// TODO(mierdin): Provide option to filter by type
		Routing(
			GET(""),
		)
		Description("Retrieve all ToDD objects.")
		Response(OK, CollectionOf(Object))
	})

	Action("get", func() {
		Routing(
			GET("/:label"),
		)
		Description("Retrieve object with given label")
		Params(func() {
			Param("label", String, "Object Label", func() {
				Minimum(1)
			})
		})
		Response(OK, Object)
		Response(NotFound)
		Response(BadRequest, ErrorMedia)
	})

	Action("create", func() {
		Routing(
			POST(""),
		)
		Description("Create new object")
		Payload(func() {
			Member("name")
			Required("name")
			// TODO(mierdin): Add label, type, and spec?
		})
		Response(Created)
		Response(BadRequest, ErrorMedia)
	})

	Action("delete", func() {
		Routing(
			DELETE("/:label"),
		)
		Params(func() {
			Param("label", String, "Object Label")
		})
		Response(NoContent)
		Response(NotFound)
		Response(BadRequest, ErrorMedia)
	})
})
