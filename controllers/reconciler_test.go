package controllers

import (
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"testing"
)

func TestGetChildSecretName(t *testing.T) {
	testStringEquality(t, getChildSecretName("mycontroller", "somesecret"),"kuberty-mycontroller-somesecret")
	testStringEquality(t, getChildSecretName("test-controller", "test-secret-2"),"kuberty-test-controller-test-secret-2")
	testStringInequality(t, getChildSecretName("test-controller-", "test-secret-2"),"kuberty-test-controller-test-secret-2")
}

func testStringEquality(t *testing.T, actual string, expected string) {
	if !reflect.DeepEqual(expected, actual) {
		t.Error("Expectation", expected, "does not equal actual", actual)
	}
}

func testStringInequality(t *testing.T, actual string, expected string) {
	if reflect.DeepEqual(expected, actual) {
		t.Error("Expectation", expected, "does not equal actual", actual)
	}
}

func TestNamespacedName(t *testing.T) {
	testNamespacedNameEquality(t, namespacedName("testname"), types.NamespacedName{Name: "testname", Namespace: ""})
	testNamespacedNameEquality(t, namespacedName("somenamespace/testname"), types.NamespacedName{Name: "testname", Namespace: "somenamespace"})
	testNamespacedNameEquality(t, namespacedName("some-namespace/test-name1"), types.NamespacedName{Name: "test-name1", Namespace: "some-namespace"})
	testNamespacedNameInequality(t, namespacedName("somenamespace/test-name1"), types.NamespacedName{Name: "testname", Namespace: "somenamespace"})
	testNamespacedNameEquality(t, namespacedName("some-namespace2/test-name1"), types.NamespacedName{Name: "test-name1", Namespace: "some-namespace2"})
	testNamespacedNameInequality(t, namespacedName("testname"), types.NamespacedName{Name: "testname", Namespace: "nonexistent"})
}

func testNamespacedNameEquality(t *testing.T, actual types.NamespacedName, expected types.NamespacedName) {
	if !reflect.DeepEqual(expected, actual) {
		t.Error("Expectation", expected, "does not equal actual", actual)
	}
}

func testNamespacedNameInequality(t *testing.T, actual types.NamespacedName, expected types.NamespacedName) {
	if reflect.DeepEqual(expected, actual) {
		t.Error("Expectation", expected, "does not equal actual", actual)
	}
}