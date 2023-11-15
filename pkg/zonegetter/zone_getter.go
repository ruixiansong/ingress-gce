/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package zonegetter

import (
	"errors"
	"fmt"
	"regexp"

	api_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/ingress-gce/pkg/annotations"
	"k8s.io/ingress-gce/pkg/utils"
	"k8s.io/klog/v2"
)

// ZoneGetter implements ZoneGetter interface
type ZoneGetter struct {
	NodeInformer cache.SharedIndexInformer
}

var ErrProviderIDNotFound = errors.New("providerID not found")
var ErrSplitProviderID = errors.New("error splitting providerID")
var ErrNodeNotFound = errors.New("node not found")

// GetZoneForNode returns the zone for a given node by looking up its zone label.
func (z *ZoneGetter) GetZoneForNode(name string) (string, error) {
	nodeLister := z.NodeInformer.GetIndexer()
	node, err := listers.NewNodeLister(nodeLister).Get(name)
	if err != nil {
		return "", fmt.Errorf("%w: failed to get node %q", ErrNodeNotFound, name)
	}
	if node.Spec.ProviderID == "" {
		return "", ErrProviderIDNotFound
	}
	zone, err := getZone(node)
	if err != nil {
		return "", err
	}
	return zone, nil
}

// ListZones returns a list of zones containing nodes that satisfy the given predicate.
func (z *ZoneGetter) ListZones(predicate utils.NodeConditionPredicate) ([]string, error) {
	nodeLister := z.NodeInformer.GetIndexer()
	return z.listZones(listers.NewNodeLister(nodeLister), predicate)
}

func (z *ZoneGetter) listZones(lister listers.NodeLister, predicate utils.NodeConditionPredicate) ([]string, error) {
	zones := sets.String{}
	nodes, err := utils.ListWithPredicate(lister, predicate)
	if err != nil {
		return zones.List(), err
	}
	for _, n := range nodes {
		zone, err := getZone(n)
		if err == nil {
			zones.Insert(zone)
		}
	}
	return zones.List(), nil
}

func getZone(node *api_v1.Node) (string, error) {
	zone, err := getZoneByProviderID(node.Spec.ProviderID)
	if err != nil {
		klog.Errorf("Failed to get zone information from node %q: %v", node.Name, err)
		return annotations.DefaultZone, err
	}
	return zone, nil
}

// getZoneByProviderID gets zone information from node provider id.
// A providerID is build out of '${ProviderName}://${project-id}/${zone}/${instance-name}'
func getZoneByProviderID(providerID string) (string, error) {
	var providerIDRE = regexp.MustCompile(`^` + "gce" + `://([^/]+)/([^/]+)/([^/]+)$`)
	matches := providerIDRE.FindStringSubmatch(providerID)
	if len(matches) != 4 {
		return "", fmt.Errorf("%w: providerID %q is not in valid format", ErrSplitProviderID, providerID)
	}
	return matches[2], nil
}
