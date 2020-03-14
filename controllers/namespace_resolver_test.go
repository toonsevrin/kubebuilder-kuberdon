package controllers

import (
	"github.com/kuberty/kuberdon/api/v1beta1"
	"reflect"
	"testing"
)

func TestDefaultResolver(t *testing.T) {
	resolver := DefaultNamespaceResolver{}
	res1, _ := resolver.resolve(
		[]string{"namespace1", "namespace2"},
		[]v1beta1.NamespaceFilter{v1beta1.NamespaceFilter{Name: "*"}})
	stringArraysEqual(t, res1, []string{"namespace1", "namespace2"})

	res2, _ := resolver.resolve(
		[]string{"namespace1", "namespace2"},
		[]v1beta1.NamespaceFilter{v1beta1.NamespaceFilter{Name: "*1"}})
	stringArraysEqual(t, res2, []string{"namespace1"})

	res3, _ := resolver.resolve(
		[]string{"namespace1", "namespace2"},
		[]v1beta1.NamespaceFilter{v1beta1.NamespaceFilter{Name: ".*2"}})
	stringArraysEqual(t, res3, []string{"namespace2"})

	res4, _ := resolver.resolve(
		[]string{"asdnamespace1", "bdsnamespace2", "asd3"},
		[]v1beta1.NamespaceFilter{v1beta1.NamespaceFilter{Name: "*namespace*"}})
	stringArraysEqual(t, res4, []string{"asdnamespace1", "bdsnamespace2"})
}

func stringArraysEqual(t *testing.T, actual []string, expected []string) {
	if !reflect.DeepEqual(expected, actual) {
		t.Error("Expectation", expected, "does not equal actual", actual)
	}
}

func TestFilterToRegex(t *testing.T) {
	testStringEquality(t, filterToRegex("*test"), ".*test")
	testStringEquality(t, filterToRegex(".*test"), ".*test")
	testStringEquality(t, filterToRegex("*test*"), ".*test.*")
	testStringEquality(t, filterToRegex("test*"), "test.*")
	testStringEquality(t, filterToRegex("test.*"), "test.*")
	testStringEquality(t, filterToRegex("tes.*t.*"), "tes.*t.*")
	testStringEquality(t, filterToRegex("tes.*t"), "tes.*t")
	testStringEquality(t, filterToRegex("*a*"), ".*a.*")
	testStringEquality(t, filterToRegex("**"), ".*.*")
	testStringEquality(t, filterToRegex("*2"), ".*2")
}
