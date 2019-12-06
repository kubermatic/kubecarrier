# Repository Organization

## Overview
Here is an overview of the KubeCarrier repository organization:
```
├── README.md, LICENSE, OWNERS, etc.
├── Makefile
├── cmd
│   ├── {anchor, manager, operator, tender}
│   │   └── main.go
├── config
│   ├── dockerfiles
│   │   ├── {manager, operator, tender, test}.Dockerfile
│   ├── internal
│   │   ├── {manager, tender}
│   └── operator
├── docs
│   └── repository-organization.md
├── hack
├── pkg
│   ├── {anchor, manager, operator, tender}
│   │   └── internal
│   ├── apis
│   ├── internal
└── test
```

*cmd* contains the `main.go` for every KubeCarrier component.

*config* contains:
- *dockerfiles*: dockerfiles for building docker images for KubeCarrier components and test image.
- *operator*: configuration (CRD, Webhook, RBAC, etc) of KubeCarrier operator.
- *internal*: configuration that used within KubeCarrier's anchor CLI and the KubeCarrier Operator to bootstrap and reconcile KubeCarrier installation, it's not meant for direct use.

*docs* provides useful documents of KubeCarrier project.

*hack* contains some useful scripts.

*pkg* contains:
- *apis*: KubeCarrier APIs.
- *interal*: internal packages used across components.
- *<component>*: source code for KubeCarrier components.
    - *internal*: internal packages for every component (controller, etc).

*test* contains KubeCarrier e2e tests.
