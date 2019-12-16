# Repository Organization

## Overview
Here is an overview of the KubeCarrier repository organization:
```
├── cmd
│   ├── {anchor, manager, operator, tender}
│   │   └── main.go
├── config
│   ├── dockerfiles
│   ├── internal
│   │   ├── {manager, tender}
│   └── operator
├── docs
├── hack
├── pkg
│   ├── {anchor, manager, operator, tender}
│   │   └── internal
│   ├── apis
│   │   ├── core
│   │   ├── catalog
│   ├── internal
└── test
```

**cmd** contains the `main.go` for every KubeCarrier component.

**config** contains:
- **dockerfiles**: dockerfiles for building docker images for KubeCarrier components and test image.
- **operator**: configuration (CRD, Webhook, RBAC, etc) of KubeCarrier operator.
- **internal**: configuration that used within KubeCarrier's anchor CLI and the KubeCarrier Operator to bootstrap and reconcile KubeCarrier installation, it's not meant for direct use.

**docs** contains the documentation of the KubeCarrier project.

**hack** contains some useful scripts.

**pkg** contains:
- **apis**: KubeCarrier APIs.
    - **core**: kubecarrier.io api group
    - **catalog**: catalog.kubecarrier.io api group
- **internal**: internal packages used across components.
- `**<component>**`: source code for KubeCarrier components.
    - **internal**: internal packages for every component (controller, etc).

**test** contains KubeCarrier e2e tests.
