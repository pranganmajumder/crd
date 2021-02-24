#!/bin/bash

/home/prangan/go/src/k8s.io/code-generator/generate-groups.sh all \
  github.com/pranganmajumder/custom-crd/pkg/client github.com/pranganmajumder/custom-crd/pkg/apis \
  demo.com:v1alpha1
