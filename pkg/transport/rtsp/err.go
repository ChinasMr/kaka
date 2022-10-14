package rtsp

func ErrInternal(res Response) Response {
	res.SetStatus("Internal Server Error")
	res.SetCode(500)
	return res
}

func ErrUnsupportedTransport(res Response) Response {
	res.SetStatus("Unsupported Transport")
	res.SetCode(461)
	return res
}

func ErrMethodNotAllowed(res Response) Response {
	res.SetStatus("Method Not Allowed")
	res.SetCode(405)
	return res
}

func ErrMethodNotValidINThisState(res Response) Response {
	res.SetStatus("Method Not Valid in This State")
	res.SetCode(455)
	return res
}
