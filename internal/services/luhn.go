package services

// LuhnService сервис для проверки номеров заказов
type LuhnService struct{}

// NewLuhnService создает новый экземпляр LuhnService
func NewLuhnService() *LuhnService {
	return &LuhnService{}
}

// Validate проверяет номер заказа с помощью алгоритма Луна
func (l *LuhnService) Validate(number string) bool {
	if number == "" {
		return false
	}

	// Проверяем, что строка состоит только из цифр
	for _, char := range number {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Алгоритм Луна
	sum := 0
	alternate := false

	// Идем справа налево
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
