#!/bin/bash

/home/prangan/go/src/k8s.io/code-generator/generate-groups.sh all \
  github.com/pranganmajumder/crd/pkg/client github.com/pranganmajumder/crd/pkg/apis \
  appscode.com:v1alpha1
