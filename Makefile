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
	cd auth && docker build --tag auth:$(VERSION) . 
	cd account && docker build --tag account:$(VERSION) . 
	cd sociallist && docker build --tag sociallist:$(VERSION) . 
	cd example/tictactoe && docker build --tag tictactoe:$(VERSION) . 
	docker tag auth:$(VERSION) auth:latest
	docker tag account:$(VERSION) account:latest
	docker tag sociallist:$(VERSION) sociallist:latest
	docker tag tictactoe:$(VERSION) tictactoe:latest
	docker tag auth:$(VERSION) k3d-myregistry.localhost:12345/auth:latest
	docker tag account:$(VERSION) k3d-myregistry.localhost:12345/account:latest
	docker tag sociallist:$(VERSION) k3d-myregistry.localhost:12345/sociallist:latest
	docker tag tictactoe:$(VERSION) k3d-myregistry.localhost:12345/tictactoe:latest
	docker push k3d-myregistry.localhost:12345/auth:latest
	docker push k3d-myregistry.localhost:12345/account:latest
	docker push k3d-myregistry.localhost:12345/sociallist:latest
	docker push k3d-myregistry.localhost:12345/tictactoe:latest

build-auth:
	cd auth && docker build --tag k3d-myregistry.localhost:12345/auth:latest .

build-account:
	cd account && docker build --tag k3d-myregistry.localhost:12345/account:latest .

build-sociallist:
	cd sociallist && docker build --tag k3d-myregistry.localhost:12345/sociallist:latest .

build-tictactoe:
	cd example/tictactoe && docker build --tag k3d-myregistry.localhost:12345/tictactoe:latest .

build: build-auth build-account build-sociallist build-tictactoe
	@echo Build done

deploy-auth: 
	docker push k3d-myregistry.localhost:12345/auth:latest

deploy-account: 
	docker push k3d-myregistry.localhost:12345/account:latest

deploy-sociallist: 
	docker push k3d-myregistry.localhost:12345/sociallist:latest

deploy-tictactoe: 
	docker push k3d-myregistry.localhost:12345/tictactoe:latest

deploy: build deploy-auth deploy-account deploy-sociallist deploy-tictactoe
	kubectl delete -f grapevine.yaml
	kubectl create -f grapevine.yaml
	@echo Deploy done

