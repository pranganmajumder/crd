CONTROLLER_GEN=/home/prangan/go/bin/controller-gen
all: manifests
.PHONY:manifests

# Generate manifests for CRDs
manifests:
	$(CONTROLLER_GEN) crd:trivialVersions=true paths="./..." output:crd:artifacts:config=config/crd/bases
