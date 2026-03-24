// pkg/eventbus/eventbus.go
package eventbus

import (
	evbus "github.com/asaskevich/EventBus"
)

// EventType 定义事件类型
type EventType string

const (
	// AlertTriggered 告警触发事件
	AlertTriggered EventType = "alert:triggered"
	// AlertRecovered 告警恢复事件
	AlertRecovered EventType = "alert:recovered"
	// NotificationSent 通知发送事件
	NotificationSent EventType = "notification:sent"
)

// Event 事件基础接口
type Event interface {
	GetType() EventType
}

// AlertEvent 告警事件
type AlertEvent struct {
	Type          EventType
	RuleID        uint
	DeviceID      string
	Parameters    interface{}
	IsManualClose bool // 是否为手动关闭
}

func (e AlertEvent) GetType() EventType {
	return e.Type
}

// EventBus 事件总线包装器
type EventBus struct {
	bus evbus.Bus
}

// NewEventBus 创建新的事件总线
func NewEventBus() *EventBus {
	return &EventBus{
		bus: evbus.New(),
	}
}

// Publish 发布事件
func (eb *EventBus) Publish(event Event) {
	eventType := string(event.GetType())
	eb.bus.Publish(eventType, event)
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(eventType EventType, handler func(Event)) error {
	return eb.bus.Subscribe(string(eventType), handler)
}

// SubscribeAsync 异步订阅事件
func (eb *EventBus) SubscribeAsync(eventType EventType, handler func(Event), transactional bool) error {
	return eb.bus.SubscribeAsync(string(eventType), handler, transactional)
}

// Unsubscribe 取消订阅
func (eb *EventBus) Unsubscribe(eventType EventType, handler func(Event)) error {
	return eb.bus.Unsubscribe(string(eventType), handler)
}

// WaitAsync 等待所有异步事件处理完成
func (eb *EventBus) WaitAsync() {
	eb.bus.WaitAsync()
}
