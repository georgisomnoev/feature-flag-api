package service

type Component interface {
	Name() string
	Check() error
}

type Service struct {
	components []Component
}

func NewService(components []Component) *Service {
	return &Service{components: components}
}

func (h *Service) Status() (map[string]string, bool) {
	status := make(map[string]string)
	allHealthy := true

	for _, c := range h.components {
		if err := c.Check(); err != nil {
			status[c.Name()] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			status[c.Name()] = "ok"
		}
	}

	return status, allHealthy
}
