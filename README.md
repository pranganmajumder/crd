
###### Workflow:
* create directory and file like that
   * Note: not custom-crd, it will be crd  
![scafolding](images/scaffold_directory.png)
    
* run `go mod init && go mod tidy && go mod vendor`
* `cd hack/`
* change mod of file update-code.sh `chmod +x update-codegen.sh`
* `./update-codegen.sh`
    * it'll generate deepcopy funcs inside v1alpha1 folder & 'clientset , listers , informers ' inside auto generated client folder
