/*
   ToDD response - upload test data

    Copyright 2016 Matt Oswalt. Use or modification of this
    source code is governed by the license provided here:
    https://github.com/toddproject/todd/blob/master/LICENSE
*/

package responses

// UploadTestDataResponse defines this particular response.
type UploadTestDataResponse struct {
	BaseResponse
	TestUuid string `json:"TestUuid"`
	TestData string `json:"status"`
}
