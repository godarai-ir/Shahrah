# Shahrah Local Kubernetes Manifests

These manifests are a Kubernetes conversion of `infra/compose/local.yml`.

## Setup

Create a Project in arvan-colud(e.g sharah). this name will be used as NS name for k8s cluster.

Current NS: `shahrah-local`

Download KubeConfig from Arvan Pannel. 

If you use Windows OS, paste this file in this dir: %USERPROFILE%/.kube/config


## Login to cluster

```bash
kubectl login
user: delpak
pass: 
```

## Verifiy kubectl configuration
```
kubectl cluster-info
```

## Apply

```bash
kubectl apply -k infra/k8s/local
```

## Container registry
`registry-def882af9a-shahrah-local.apps.ir-central1.arvancaas.ir`
volume: 10GB

### Container registry account
user: delpak

pass: ******

## Build Images First

The following images come from compose `build:` entries and must exist in your cluster runtime:

- `shahrah/control-plane:latest`
- `shahrah/model-discovery:latest`
- `shahrah/model-discovery-migrate:latest`
- `shahrah/model-discovery-web:latest`
- `shahrah/support-admin:latest`
- `compose-tunnel-client:latest`
- `shahrah/tunnel-server:latest`
- `shahrah/silk-road:latest`
- `shahrah/goose:latest`

Example build commands from repo root:

```bash
docker build -t shahrah/control-plane:local -f services/control-plane/Dockerfile .
docker build -t registry-def882af9a-shahrah-local.apps.ir-central1.arvancaas.ir/shahrah-local/model-discovery:latest -f apps/model-discovery/Dockerfile . --push
docker build -t shahrah/model-discovery-migrate:local --target migrate -f apps/model-discovery/Dockerfile .
docker build -t shahrah/model-discovery-web:local -f apps/model-discovery-web/Dockerfile .
docker build -t shahrah/support-admin-migrate:local --target migrate -f apps/support-admin/Dockerfile .
docker build -t shahrah/support-admin:local -f apps/support-admin/Dockerfile .
docker build -t registry-def882af9a-shahrah-local.apps.ir-central1.arvancaas.ir/shahrah-local/tunnel-client:latest -f services/tunnel-client/Dockerfile services/tunnel-client --push
docker build -t shahrah/tunnel-server:local -f services/tunnel-server/Dockerfile services/tunnel-server
docker build -t shahrah/silk-road:local -f services/silk-road/Dockerfile .
docker build -t shahrah/goose:local -f infra/compose/images/goose.Dockerfile .
```

# 2) Tag and push all locally-built app images
docker tag shahrah/control-plane:local          $REG/$NS/control-plane:local
docker tag shahrah/model-discovery:local        $REG/$NS/model-discovery:local
docker tag shahrah/model-discovery-migrate:local $REG/$NS/model-discovery-migrate:local
docker tag shahrah/model-discovery-web:local    $REG/$NS/model-discovery-web:local
docker tag shahrah/support-admin:local          $REG/$NS/support-admin:local
docker tag shahrah/tunnel-client:local          $REG/$NS/tunnel-client:local
docker tag shahrah/tunnel-server:local          $REG/$NS/tunnel-server:local
docker tag shahrah/silk-road:local              $REG/$NS/silk-road:local
docker tag shahrah/goose:local                  $REG/$NS/goose:local

docker push $REG/$NS/control-plane:local
docker push $REG/$NS/model-discovery:local
docker push $REG/$NS/model-discovery-migrate:local
docker push $REG/$NS/model-discovery-web:local
docker push $REG/$NS/support-admin:local
docker push $REG/$NS/tunnel-client:local
docker push $REG/$NS/tunnel-server:local
docker push $REG/$NS/silk-road:local
docker push $REG/$NS/goose:local
## Access Services

All services are `ClusterIP` by default. Use port-forward when needed:

```bash
kubectl -n shahrah-local port-forward svc/control-plane 4000:4000
kubectl -n shahrah-local port-forward svc/model-discovery 3002:3002
kubectl -n shahrah-local port-forward svc/model-discovery-web 3003:3000
kubectl -n shahrah-local port-forward svc/support-admin 3001:3000
kubectl -n shahrah-local port-forward svc/mailpit 8025:8025
kubectl -n shahrah-local port-forward svc/pgadmin 5050:80
```

## Notes

- Compose `depends_on` is represented with `initContainers` that wait for dependent TCP ports.
- Compose named volumes are represented with PVCs.
- Compose migration containers are represented as Kubernetes `Job`s.
