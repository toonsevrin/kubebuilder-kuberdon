package controllers

import (
	"context"
	kuberdonv1beta1 "github.com/kuberty/kuberdon/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type State struct {
	MasterSecret v1.Secret
	Namespaces   map[string]*NamespaceState // key is the namespace, value is the child state in that namespace
}

// A small struct we use to reduce the amount of parameters in our getActualState and getDesiredState
type Reconciliation struct {
	Client     client.Client
	Context    context.Context
	Registry   kuberdonv1beta1.Registry
	Namespaces v1.NamespaceList
}

type NamespaceState struct {
	ChildSecret     v1.Secret         // Not always present
	ServiceAcccount v1.ServiceAccount // Should always be present
}

type States struct {
	Actual  State
	Desired State
}

func (s Reconciliation) getActualAndDesiredState(resolver NamespaceResolver) (States, error) {
	actualState, err := s.getActualState()
	if err != nil {
		return States{}, err
	}
	//fetch the desired state
	desiredState, err := s.getDesiredState(actualState, resolver)
	if err != nil {
		return States{}, err
	}
	return States{Actual: actualState, Desired: desiredState}, nil
}
