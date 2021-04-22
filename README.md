[![Go Report Card](https://goreportcard.com/badge/github.com/pranganmajumder/crd)](https://goreportcard.com/report/github.com/pranganmajumder/crd)
###### Workflow:
* create directory and file like that
   * Note: not custom-crd, it will be crd  
![scafolding](images/scaffold_directory.png)
    
* run `go mod init && go mod tidy && go mod vendor`
* `cd hack/`
* change mod of file update-code.sh `chmod +x update-codegen.sh`
* `./update-codegen.sh`
    * it'll generate deepcopy funcs inside v1alpha1 folder & 'clientset , listers , informers ' inside auto generated client folder
  
* now run `make` after putting necessary rules into `Makefile`
  * it'll generate `crd` inside auto generated `config/crd/bases/` directory named `appscode.com_apployments.yaml`

## Now Run your custorm Controller

```sh
# assumes you have a working kubeconfig, not required if operating in-cluster
go build -o controller-bin .
./controller-bin -kubeconfig=$HOME/.kube/config

# create a CustomResourceDefinition
kubectl create -f config/crd/bases/appscode.com_apployments.yaml

# create a custom resource of type Foo
kubectl create -f apps-object/apployment.yaml

# check deployments created through the custom resource
kubectl get deployments
kubectl get apployment
kubectl get pod
kubectl get svc
# it'll delete our number of desired replica but after a while our desired number of replica will be regenerated through custom controller 
kubectl delete deployment testing 
```
##### Resources:
* https://www.youtube.com/watch?v=XaF3JvzBnjM
* https://www.youtube.com/watch?v=41tflsbaZPs
* https://www.youtube.com/watch?v=a005Qlx11qc
* https://engineering.bitnami.com/articles/a-deep-dive-into-kubernetes-controllers.html
* https://github.com/kubernetes/client-go/blob/master/examples/workqueue/main.go


##### Error:
* `package github.com/googleapis/gnostic/OpenAPIv2: cannot find package` for this run `go get github.com/googleapis/gnostic@v0.4.0` inside directory `crd`
