package main

type Response struct {
	Success bool   		`json:"success"`
	Message string 		`json:"message"`
	Data	interface{} `json:"data"`
}

//add a function on the struct that returns a default response

func DefaultResponse() Response {
	return Response{
		Success: false,
		Message: "nothing to see here, go away",
		Data:    map[string]interface{}{},
	}
}

func ServerErrorResponse(message string) Response {
	return Response{
		Success: false,
		Message: "internal server error: " + message,
		Data:    map[string]interface{}{},
	}
}