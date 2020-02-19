# End-to-End testing

The e2e testing infrastructure is powered by Drone and Kubernetes Kind.

### CI workflow

* download go modules
* run unit tests
* build container
* install kubectl, Helm v3 and Kubernetes Kind CLIs
* create local Kubernetes cluster with kind
* load image onto the local cluster
* deploy with Helm
* set the image to the locally built one
* run Helm tests