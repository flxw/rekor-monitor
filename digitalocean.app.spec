alerts:
- rule: DEPLOYMENT_FAILED
- rule: DOMAIN_FAILED
databases:
- engine: PG
  name: db
  num_nodes: 1
  size: basic-xxs
  version: "12"
ingress:
  rules:
  - component:
      name: grafana-grafana-enterprise
    match:
      path:
        prefix: /
name: lobster-app
region: fra
services:
- http_port: 3000
  image:
    registry: grafana
    registry_type: DOCKER_HUB
    repository: grafana-enterprise
    tag: latest
  instance_count: 1
  instance_size_slug: basic-xxs
  name: grafana-grafana-enterprise
workers:
- image:
    registry: library
    registry_type: DOCKER_HUB
    repository: hello-world
    tag: latest
  instance_count: 1
  instance_size_slug: basic-xxs
  name: hello-world

