package specs

type Response struct {
	Description string
	MediaType   string
	Value       interface{}
}

type Body struct {
	MediaType string
	Value     interface{}
}

type Endpoint[T interface{}] struct {
	OperationID string
	Title       string
	Description string

	Deprecated bool
	Tags       []string

	Handler T

	Protocol string
	Method   string
	Path     string
	Status   int

	Parameters interface{}
	Query      interface{}

	Payload  []Body
	Response map[int]Response
}
