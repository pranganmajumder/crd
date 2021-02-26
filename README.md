
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
  
* `kc apply -f appscode.com_apployments.yaml`
  * it'll generate the crd named `apployments.appscode.com` now create your custom yaml resource file 
  * apply custom object which you created apps-object/apployment.yaml manually `kc apply -f apployment.yaml` . 
  * It'll create apployment , run `kc get apployment`
