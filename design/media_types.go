package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

// Object is the ToDD object resource media type.
var Object = MediaType("application/vnd.account+json", func() {
	Description("A ToDD object (group, testrun, etc)")
	Attributes(func() {
		Attribute("label", String, "The label for the object", func() {
			Example(1) // WTF is this?
		})
		Attribute("type", String, "Type of object (group, testrun, etc)", func() {
			Example("/accounts/1") // WTF is this?
		})
		Attribute("spec", String, "Details of object", func() {
			Example("/accounts/1") // WTF is this?
		})
		// Attribute("created_by", String, "Email of account owner", func() {
		// 	Format("email")
		// 	Example("me@goa.design")
		// })
		Required("label", "type")
	})

	View("default", func() {
		Attribute("label")
		Attribute("type")
		Attribute("spec")
	})

	// View("tiny", func() {
	// 	Description("tiny is the view used to list accounts")
	// 	Attribute("id")
	// 	Attribute("href")
	// 	Attribute("name")
	// })

	// View("link", func() {
	// 	Attribute("id")
	// 	Attribute("href")
	// })
})

// // Bottle is the bottle resource media type.
// var Bottle = MediaType("application/vnd.bottle+json", func() {
// 	Description("A bottle of wine")
// 	Reference(BottlePayload)
// 	Attributes(func() {
// 		Attribute("id", Integer, "ID of bottle", func() {
// 			Example(1)
// 		})
// 		Attribute("href", String, "API href of bottle", func() {
// 			Example("/accounts/1/bottles/1")
// 		})
// 		Attribute("rating", Integer, "Rating of bottle between 1 and 5", func() {
// 			Minimum(1)
// 			Maximum(5)
// 		})
// 		Attribute("account", Account, "Account that owns bottle")
// 		Attribute("created_at", DateTime, "Date of creation")
// 		Attribute("updated_at", DateTime, "Date of last update")
// 		// Attributes below inherit from the base type
// 		Attribute("name")
// 		Attribute("vineyard")
// 		Attribute("varietal")
// 		Attribute("vintage")
// 		Attribute("color")
// 		Attribute("sweetness")
// 		Attribute("country")
// 		Attribute("region")
// 		Attribute("review")

// 		Required("id", "href", "name", "vineyard", "varietal", "vintage", "color")
// 		Required("created_at")
// 	})

// 	Links(func() {
// 		Link("account")
// 	})

// 	View("default", func() {
// 		Attribute("id")
// 		Attribute("href")
// 		Attribute("name")
// 		Attribute("rating")
// 		Attribute("vineyard")
// 		Attribute("varietal")
// 		Attribute("vintage")
// 		Attribute("account", func() {
// 			View("tiny")
// 		})
// 		Attribute("links")
// 	})

// 	View("tiny", func() {
// 		Attribute("id")
// 		Attribute("href")
// 		Attribute("name")
// 		Attribute("rating")
// 		Attribute("links")
// 	})

// 	View("full", func() {
// 		Attribute("id")
// 		Attribute("href")
// 		Attribute("name")
// 		Attribute("account")
// 		Attribute("rating")
// 		Attribute("vineyard")
// 		Attribute("varietal")
// 		Attribute("vintage")
// 		Attribute("color")
// 		Attribute("sweetness")
// 		Attribute("country")
// 		Attribute("region")
// 		Attribute("review")
// 		Attribute("created_at")
// 		Attribute("updated_at")
// 		Attribute("links")
// 	})
// })
