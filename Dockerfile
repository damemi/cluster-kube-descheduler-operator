FROM registry.svc.ci.openshift.org/openshift/release:golang-1.10 AS builder
WORKDIR /go/src/github.com/openshift/descheduler-operator
COPY . .
RUN go build -o descheduler-operator ./cmd/manager

FROM registry.svc.ci.openshift.org/openshift/origin-v4.0:base
COPY --from=builder /go/src/github.com/openshift/descheduler-operator/descheduler-operator /usr/bin/
LABEL io.k8s.display-name="OpenShift Descheduler Operator" \
      io.k8s.description="This is a component of OpenShift and manages the descheduler" \
      io.openshift.tags="openshift,descheduler-operator" \
      maintainer="AOS pod team, <aos-pod@redhat.com>"