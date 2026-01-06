package utils

type Responser[Response any] interface {
	ToResponse() Response
}

func ToResponse[Entity Responser[Response], Response any](items []Entity) []Response {
	result := make([]Response, 0, len(items))
	for _, item := range items {
		result = append(result, item.ToResponse())
	}
	return result
}
