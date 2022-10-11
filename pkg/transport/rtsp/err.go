package rtsp

func ErrNotFound(res Response) {
	res.SetStatus("Not Found")
	res.SetCode(404)
}

func ErrInternal(res Response) {
	res.SetStatus("Internal Server Error")
	res.SetCode(500)
}

func ErrUnsupportedTransport(res Response) {
	res.SetStatus("Unsupported Transport")
	res.SetCode(461)
}

func ErrMethodNotAllowed(res Response) {
	res.SetStatus("Method Not Allowed")
	res.SetCode(405)
}

func ErrMethodNotValidINThisState(res Response) {
	res.SetStatus("Method Not Valid in This State")
	res.SetCode(455)
}
