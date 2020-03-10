## Documentation
This is a specification of the inner workings of Kuberdon.

### CRD controller
The controller is managed by [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), a framework that removes a lot of the friction of traditional controller development.

### Namespace matching
Namespaces are matched according to the namespace selectors in the registry spec. This selector can be defined as a regex.

If you for example use `*-blabla`, `blabla-*` or `*` for the name, Kuberdon maps it to `.*-blabla`, `blabla-.*` and `.*` respectively.


#### Priority order
The first rules take priority when they are valid on a namespace.

**Good priority order example:**
```yaml
  namespaces:
  - name: "kube-system"
    exclude: true
  - name: "*"
    add-automatically: true
```

**Bad priority order example:**
```yaml
  namespaces:
  - name: "*"
    add-automatically: true
  - name: "kube-system"
    exclude: true
```
Here the second rule is always ignored, even for the kube-system namespace, as it already matches the first rule.
### Collission-avoidance
To avoid collissions, kuberdon prefixes all deployed secrets with 'kuberdon-'. 

Kuberdon also sets the ownerReferences to the Kuberdon Registry. If a kuberdon- prefixed secret already exists with a dfferent owner, kuberdon will display this in the status

### Garbage collection
For this to work, Registry objects have to be cluster scoped.
```yaml
ownerReferences:
  - apiVersion: kuberdon.kuberty.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: Registry
    name: kuberty-gitlab-read
    uid: 24c17568-daa9-4cbb-b121-f5bd42dc703a
```
