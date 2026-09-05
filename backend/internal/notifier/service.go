package notifier

type Service struct {
	telegram Sender
	vk       Sender
}

func NewService(
	telegram Sender,
	vk Sender,
) *Service {
	return &Service{
		telegram: telegram,
		vk:       vk,
	}
}
