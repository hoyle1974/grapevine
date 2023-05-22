#!/bin/bash

up()
{
	kubectl apply -f https://k8s.io/examples/admin/dns/dnsutils.yaml
	echo "Try: kubectl exec -i -t dnsutils -- nslookup kubernetes.default"
}

down() {
	kubectl delete -f https://k8s.io/examples/admin/dns/dnsutils.yaml
}

.
. ./base.sh $1
