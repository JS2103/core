// Copyright 2021 The Cloud Robotics Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/googlecloudrobotics/core/src/go/pkg/apis/registry/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// RobotLister helps list Robots.
// All objects returned here must be treated as read-only.
type RobotLister interface {
	// List lists all Robots in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Robot, err error)
	// Robots returns an object that can list and get Robots.
	Robots(namespace string) RobotNamespaceLister
	RobotListerExpansion
}

// robotLister implements the RobotLister interface.
type robotLister struct {
	indexer cache.Indexer
}

// NewRobotLister returns a new RobotLister.
func NewRobotLister(indexer cache.Indexer) RobotLister {
	return &robotLister{indexer: indexer}
}

// List lists all Robots in the indexer.
func (s *robotLister) List(selector labels.Selector) (ret []*v1alpha1.Robot, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Robot))
	})
	return ret, err
}

// Robots returns an object that can list and get Robots.
func (s *robotLister) Robots(namespace string) RobotNamespaceLister {
	return robotNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// RobotNamespaceLister helps list and get Robots.
// All objects returned here must be treated as read-only.
type RobotNamespaceLister interface {
	// List lists all Robots in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Robot, err error)
	// Get retrieves the Robot from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Robot, error)
	RobotNamespaceListerExpansion
}

// robotNamespaceLister implements the RobotNamespaceLister
// interface.
type robotNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Robots in the indexer for a given namespace.
func (s robotNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Robot, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Robot))
	})
	return ret, err
}

// Get retrieves the Robot from the indexer for a given namespace and name.
func (s robotNamespaceLister) Get(name string) (*v1alpha1.Robot, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("robot"), name)
	}
	return obj.(*v1alpha1.Robot), nil
}
