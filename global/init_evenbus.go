package global

import "github.com/asaskevich/EventBus"

var Bus EventBus.Bus

func init() {
	Bus = EventBus.New()
}
