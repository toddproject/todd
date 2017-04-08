package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = Resource("object", func() {

	DefaultMedia(Object)
	BasePath("/object")

	Action("list", func() {
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
		Response(OK)
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
		// Response(Created, "/accounts/[0-9]+")  // WTF is this?
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

// var _ = Resource("bottle", func() {

// 	DefaultMedia(Bottle)
// 	BasePath("bottles")
// 	Parent("account")

// 	Action("list", func() {
// 		Routing(
// 			GET(""),
// 		)
// 		Description("List all bottles in account optionally filtering by year")
// 		Params(func() {
// 			Param("years", ArrayOf(Integer), "Filter by years")
// 		})
// 		Response(OK, func() {
// 			Media(CollectionOf(Bottle, func() {
// 				View("default")
// 				View("tiny")
// 			}))
// 		})
// 		Response(NotFound)
// 		Response(BadRequest, ErrorMedia)
// 	})

// 	Action("show", func() {
// 		Routing(
// 			GET("/:bottleID"),
// 		)
// 		Description("Retrieve bottle with given id")
// 		Params(func() {
// 			Param("bottleID", Integer)
// 		})
// 		Response(OK)
// 		Response(NotFound)
// 		Response(BadRequest, ErrorMedia)
// 	})

// 	Action("watch", func() {
// 		Routing(
// 			GET("/:bottleID/watch"),
// 		)
// 		Scheme("ws")
// 		Description("Retrieve bottle with given id")
// 		Params(func() {
// 			Param("bottleID", Integer)
// 		})
// 		Response(SwitchingProtocols)
// 		Response(BadRequest, ErrorMedia)
// 	})

// 	Action("create", func() {
// 		Routing(
// 			POST(""),
// 		)
// 		Description("Record new bottle")
// 		Payload(BottlePayload, func() {
// 			Required("name", "vineyard", "varietal", "vintage", "color")
// 		})
// 		Response(Created, "^/accounts/[0-9]+/bottles/[0-9]+$")
// 		Response(NotFound)
// 		Response(BadRequest, ErrorMedia)
// 	})

// 	Action("update", func() {
// 		Routing(
// 			PATCH("/:bottleID"),
// 		)
// 		Params(func() {
// 			Param("bottleID", Integer)
// 		})
// 		Payload(BottlePayload)
// 		Response(NoContent)
// 		Response(NotFound)
// 		Response(BadRequest, ErrorMedia)
// 	})

// 	Action("rate", func() {
// 		Routing(
// 			PUT("/:bottleID/actions/rate"),
// 		)
// 		Params(func() {
// 			Param("bottleID", Integer)
// 		})
// 		Payload(func() {
// 			Member("rating", Integer)
// 			Required("rating")
// 		})
// 		Response(NoContent)
// 		Response(NotFound)
// 		Response(BadRequest, ErrorMedia)
// 	})

// 	Action("delete", func() {
// 		Routing(
// 			DELETE("/:bottleID"),
// 		)
// 		Params(func() {
// 			Param("bottleID", Integer)
// 		})
// 		Response(NoContent)
// 		Response(NotFound)
// 		Response(BadRequest, ErrorMedia)
// 	})
// })

// var _ = Resource("public", func() {
// 	Origin("*", func() {
// 		Methods("GET", "OPTIONS")
// 	})
// 	Files("/ui", "public/html/index.html")
// })

// var _ = Resource("js", func() {
// 	Origin("*", func() {
// 		Methods("GET", "OPTIONS")
// 	})
// 	Files("/js/*filepath", "public/js")
// })

// var _ = Resource("swagger", func() {
// 	Origin("*", func() {
// 		Methods("GET", "OPTIONS")
// 	})
// 	Files("/swagger.json", "public/swagger/swagger.json")
// })
