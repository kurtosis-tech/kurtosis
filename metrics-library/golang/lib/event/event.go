package event

import (
	"strings"
)

type Event struct {
	//Category of event (e.g. enclave, module)
	category string

	//Action performed/event name (e.g. create, load)
	action string

	//Properties' keys and values associated with the object of the action (e.g. enclave ID, module name)
	properties map[string]string
}

// WARNING: It's VERY important that this doesn't return an error, else the error will propagate and
//
//	lead to the event not even getting sent to Segment (which means we silently drop the event with
//	no chance to realize what's happening!)
func newEvent(category string, action string, properties map[string]string) *Event {
	categoryWithoutSpaces := strings.TrimSpace(category)
	actionWithoutSpaces := strings.TrimSpace(action)

	propertiesToSend := map[string]string{}
	for propertyKey, propertyValue := range properties {
		propertyKeyWithoutSpaces := strings.TrimSpace(propertyKey)
		propertiesToSend[propertyKeyWithoutSpaces] = propertyValue
	}

	event := &Event{
		category:   categoryWithoutSpaces,
		action:     actionWithoutSpaces,
		properties: propertiesToSend,
	}

	return event
}

func (event *Event) GetCategory() string {
	return event.category
}

func (event *Event) GetAction() string {
	return event.action
}

func (event *Event) GetName() string {
	return strings.Join([]string{event.category, event.action}, "-")
}

func (event *Event) GetProperties() map[string]string {
	return event.properties
}
