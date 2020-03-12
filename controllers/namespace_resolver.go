package controllers

import (
	"github.com/kuberty/kuberdon/api/v1beta1"
	"regexp"
)

type NamespaceResolver interface {
	resolve(allNamespaces []string, filter v1beta1.NamespaceFilter) []string
}

type DefaultNamespaceResolver struct {}


/**
	Messy implementation to handle the namespace resolving:
	The reason this implementation is so long is to allow GLOBs ("*test*") instead of just regex (".*test.*")
 */
func (s *DefaultNamespaceResolver) resolve(allNamespaces []string, filters []v1beta1.NamespaceFilter) ([]string, int){
	faultyRegex := 0
	deployToNamespaces := map[string]bool{}
	namespaces := []string{}
	for _, filter := range filters {
		print(filter.Name)
		regexStr := filterToRegex(filter.Name)
		exp, err := regexp.Compile(regexStr)
		if err != nil {
			faultyRegex += 1
		}else {
			for _, namespace := range allNamespaces {
				if exp.MatchString(namespace) {
					if _, ok := deployToNamespaces[namespace]; !ok {
						namespaces = append(namespaces, namespace)
						deployToNamespaces[namespace] = true
					}
				}
			}
		}
	}
	return namespaces, faultyRegex
}

//Compiles the filter into a string regexp
func filterToRegex(filter string) string {
	exp := filter
	if filter == "*" {
		exp = ".*"
	}else {
		if filter[len(filter)-1:] == "*" && filter[len(filter) -2:] != ".*" && isName(filter[1:len(filter) - 1]) {
			exp = exp[:len(filter) -1] + ".*"
		}
		if filter[0:1] == "*" && isName(filter[1:len(filter) -1]){
			exp = ".*" + exp[1:]
		}
	}
	return exp
}
var isNameRegex = regexp.MustCompile("^[a-z|A-Z|0-9|\\-|\\.]*$")
//check whether or not the str is a kubernetes name https://kubernetes.io/docs/concepts/overview/working-with-objects/names/ (note that we do allow capitalized letters here, we can delete this)
func isName(str string) bool{
	return isNameRegex.MatchString(str)
}