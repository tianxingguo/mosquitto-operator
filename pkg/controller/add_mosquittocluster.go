package controller

import (
	"mosquitto-operator/pkg/controller/mosquittocluster"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, mosquittocluster.Add)
}
