package handlers

// интерфейс методов бизнес логики
type Services interface {
}

// структура для конструктура обработчика
type Handler struct {
	service Services
}

// конструктор обработчика
func NewHandler(s Services) *Handler {
	return &Handler{
		s,
	}
}
