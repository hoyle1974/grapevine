all:
	@echo done

cluster:
	-k3d cluster delete grapevine
	-k3d registry delete k3d-myregistry.localhost
	k3d registry create myregistry.localhost --port 12345
	k3d cluster create grapevine --registry-use k3d-myregistry.localhost:12345
	kubectl rollout status deployment local-path-provisioner -n kube-system
	kubectl rollout status deployment coredns -n kube-system
	kubectl rollout status deployment metrics-server -n kube-system
	sleep 10
	echo "-------- CLUSTER READY ---------"
	helm install postgres oci://registry-1.docker.io/bitnamicharts/postgresql --values postgres-values.yaml
	kubectl wait pods -n default postgres-postgresql-0 --for condition=Ready --timeout=600s
	kubectl port-forward --namespace default svc/postgres-postgresql 5432:5432 &
	sleep 3
	PGPASSWORD="postgres" psql --host 127.0.0.1 -U postgres -d grapevine -p 5432 -f schema.sql


protos:  proto/account.proto proto/list.proto proto/auth.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/common.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/account.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/list.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/auth.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/grapevine.proto

VERSION := $(shell date +%s)

docker:
	#-docker image rm auth
	#-docker image rm k3d-myregistry.localhost:12345/auth:$(VERSION)
	cd auth && docker build --tag auth:$(VERSION) . 
	docker tag auth:$(VERSION) k3d-myregistry.localhost:12345/auth:latest
	docker push k3d-myregistry.localhost:12345/auth:latest

