# Debugging kubecarrier guide

## e2e test

```bash
make clean && make e2e-setup
```

usually the command `make e2e-test` is run, which first calls `make e2e-setup`, then `kubectl kubecarrier e2e-test run` and finishes with some post-processing. During e2e test bunch of objects are created, and subsequently deleted. This deletion could be controlled via the `--clean-up-strategy` CLI flag (`on-success`, `never`, `always`)

## guide

This is a small cheat sheet for common kubecarrier debugging operations during development.
During e2e tests the kind clusters are created with the `./hack/kind-config.yaml` and `./hack/audit.yaml` kube-apiserver audit logging

The kind created nodes as docker containers, thus the audit logs are saved under `/var/log/kube-apiserver-audit.log` inside the docker container.
They can be extracted with `docker cp kubecarrier-1-control-plane:/var/log/kube-apiserver-audit.log /tmp/audit.json` or `docker exec kubecarrier-1-control-plane tail -n +1 -f  /var/log/kube-apiserver-audit.log` for streaming version.

In the e2e-test dump the audit logs are located under `management` and `svc` folder. The logs could easily be processed using the `jq` tool. Here are few common operations:

```shell
cat audit.log | jq -C 'select(.user.username=="system:serviceaccount:kubecarrier-system:kubecarrier-operator-sa" and .stage=="ResponseComplete" and .verb=="create") | {stageTimestamp, verb, requestURI} + {"name": .objectRef?.name, "code": .responseStatus?.code} ' | less
```

container logs are also located in the e2e-test log dump. They can easily be processed like this:

```bash
cat <file.log> |cut -d ' ' -f 4- | jq 'select(.level=="error") | select(.error | contains("has been modified") | not) | "\(.ts | todate), \(.error)"'
```

the `cut -d ' ' -f 4-` drops first 3 fields which are added to the log lines by the containerd


For full jq information look at the official documentation

## SUT

There's also almighty SUT command, which replaces any kubecarrier deployment with telepresence, and creates appropriate IDE tasks for running that pod on the local host, optionally in a debugger.

`kubectl kubecarrier sut <component name>` and follow the on-screen instructions


## Various make targets

e2e test and kubecarrier component deployments IDE tasks are created via the `make generate-ide-tasks` There's also `make install-git-hooks` to regenerate ide-tasks on branch change or a new commit.

(This is the quite important for setting the proper version in the operator, since it spins up new deployments with the appropriate image tag)

For the e2e test this allows setting debugging breakpoints during e2e test execution. For the component it allows running component locally, though SUT is usually more appropriate solution.
