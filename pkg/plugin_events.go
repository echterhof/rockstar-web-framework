package pkg

import (
	"fmt"
	"sync"
	"time"
)

// eventBusImpl is the default implementation of EventBus
type eventBusImpl struct {
	mu            sync.RWMutex
	subscriptions map[string]map[string]EventHandler // event -> pluginName -> handler
	logger        Logger
}

// NewEventBus creates a new event bus
func NewEventBus(logger Logger) EventBus {
	return &eventBusImpl{
		subscriptions: make(map[string]map[string]EventHandler),
		logger:        logger,
	}
}

// Publish publishes an event to all subscribers asynchronously
func (e *eventBusImpl) Publish(eventName string, data interface{}) error {
	if eventName == "" {
		return fmt.Errorf("event name cannot be empty")
	}

	e.mu.RLock()
	subscribers, exists := e.subscriptions[eventName]
	if !exists || len(subscribers) == 0 {
		e.mu.RUnlock()
		// No subscribers, not an error
		return nil
	}

	// Create a copy of subscribers to avoid holding the lock during delivery
	handlers := make(map[string]EventHandler, len(subscribers))
	for pluginName, handler := range subscribers {
		handlers[pluginName] = handler
	}
	e.mu.RUnlock()

	// Create the event
	event := Event{
		Name:      eventName,
		Data:      data,
		Source:    "", // Source will be set by the publishing plugin context
		Timestamp: time.Now(),
	}

	// Deliver events asynchronously to all subscribers
	for pluginName, handler := range handlers {
		go e.deliverEvent(pluginName, event, handler)
	}

	return nil
}

// deliverEvent delivers an event to a single subscriber with error isolation
func (e *eventBusImpl) deliverEvent(pluginName string, event Event, handler EventHandler) {
	defer func() {
		if r := recover(); r != nil {
			if e.logger != nil {
				e.logger.Error(fmt.Sprintf("Plugin %s event handler panicked for event %s: %v", pluginName, event.Name, r))
			}
		}
	}()

	if err := handler(event); err != nil {
		if e.logger != nil {
			e.logger.Error(fmt.Sprintf("Plugin %s event handler failed for event %s: %v", pluginName, event.Name, err))
		}
	}
}

// Subscribe registers a plugin's event handler for a specific event
func (e *eventBusImpl) Subscribe(pluginName, eventName string, handler EventHandler) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if eventName == "" {
		return fmt.Errorf("event name cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("event handler cannot be nil")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.subscriptions[eventName]; !exists {
		e.subscriptions[eventName] = make(map[string]EventHandler)
	}

	// Check if already subscribed
	if _, exists := e.subscriptions[eventName][pluginName]; exists {
		return fmt.Errorf("plugin %s is already subscribed to event %s", pluginName, eventName)
	}

	e.subscriptions[eventName][pluginName] = handler

	if e.logger != nil {
		e.logger.Info(fmt.Sprintf("Plugin %s subscribed to event %s", pluginName, eventName))
	}

	return nil
}

// Unsubscribe removes a plugin's subscription to an event
func (e *eventBusImpl) Unsubscribe(pluginName, eventName string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if eventName == "" {
		return fmt.Errorf("event name cannot be empty")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	subscribers, exists := e.subscriptions[eventName]
	if !exists {
		return fmt.Errorf("no subscriptions found for event %s", eventName)
	}

	if _, exists := subscribers[pluginName]; !exists {
		return fmt.Errorf("plugin %s is not subscribed to event %s", pluginName, eventName)
	}

	delete(subscribers, pluginName)

	// Clean up empty subscription maps
	if len(subscribers) == 0 {
		delete(e.subscriptions, eventName)
	}

	if e.logger != nil {
		e.logger.Info(fmt.Sprintf("Plugin %s unsubscribed from event %s", pluginName, eventName))
	}

	return nil
}

// ListSubscriptions returns the names of all plugins subscribed to an event
func (e *eventBusImpl) ListSubscriptions(eventName string) []string {
	if eventName == "" {
		return []string{}
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	subscribers, exists := e.subscriptions[eventName]
	if !exists {
		return []string{}
	}

	names := make([]string, 0, len(subscribers))
	for pluginName := range subscribers {
		names = append(names, pluginName)
	}

	return names
}

// UnregisterAll removes all event subscriptions for a plugin
func (e *eventBusImpl) UnregisterAll(pluginName string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	removedCount := 0
	for eventName, subscribers := range e.subscriptions {
		if _, exists := subscribers[pluginName]; exists {
			delete(subscribers, pluginName)
			removedCount++

			// Clean up empty subscription maps
			if len(subscribers) == 0 {
				delete(e.subscriptions, eventName)
			}
		}
	}

	if e.logger != nil {
		e.logger.Info(fmt.Sprintf("Unregistered %d event subscriptions for plugin %s", removedCount, pluginName))
	}

	return nil
}
