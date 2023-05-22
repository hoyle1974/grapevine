#!/bin/bash

kubectl delete pod busybox 
kubectl run --image=busybox -i --tty busybox --restart=Never -- sh
