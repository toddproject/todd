package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

// Object is the ToDD object resource media type.
var Object = MediaType("application/json", func() {

	Description("A ToDD object (group, testrun, etc)")

	Attributes(func() {

		Attribute("label", String, func() {
			Description("The label for the object")
			Example("test-datacenter")
		})

		Attribute("type", String, func() {
			Description("Type of object (group, testrun, etc)")
			Enum("testrun", "group")
		})

		Attribute("spec", String, "Details of object", func() {
			// TODO(mierdin): I _think_ this can just be a string, since we're doing our own
			// parsing of this after it's marshaled into BaseObject
			Example("{}")
		})

		Required("label", "type", "spec")
	})

	View("default", func() {
		Attribute("label")
		Attribute("type")
		Attribute("spec")
	})
})
