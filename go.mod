module k8s.io/kubernetes

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/kuberty/kuberdon v0.0.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)

replace github.com/kuberty/kuberdon v0.0.0 => ./
