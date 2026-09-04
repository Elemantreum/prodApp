package events

import (
	"context"
	"log"
	"reflect"
	"sync"
)

type Bus struct {
	mu       sync.RWMutex
	handlers map[reflect.Type][]any
}

func NewBus() *Bus {
	return &Bus{handlers: make(map[reflect.Type][]any)}
}

// Subscribe регистрирует обработчик события типа T. Вызывается один раз при
// сборке зависимостей в main.go, до старта сервера — не потокобезопасен
// относительно Publish того же типа, но это и не нужно: подписки не меняются
// во время работы приложения.
func Subscribe[T any](bus *Bus, handler func(ctx context.Context, event T)) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	t := reflect.TypeOf((*T)(nil)).Elem()
	bus.handlers[t] = append(bus.handlers[t], handler)
}

// Publish синхронно вызывает всех подписчиков T в порядке регистрации.
// Паника в обработчике гасится и логируется — сбой уведомления не должен
// ронять того, кто событие опубликовал (ровно та проблема, что сейчас есть
// в AddTraining с NotifyAll). Если обработчику самому нужна асинхронность —
// он сам оборачивается в go func, шина в это не лезет.
func Publish[T any](ctx context.Context, bus *Bus, event T) {
	bus.mu.RLock()
	handlers := bus.handlers[reflect.TypeOf(event)]
	bus.mu.RUnlock()

	for _, h := range handlers {
		fn, ok := h.(func(ctx context.Context, event T))
		if !ok {
			continue
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("events: handler for %T panicked: %v", event, r)
				}
			}()
			fn(ctx, event)
		}()
	}
}
