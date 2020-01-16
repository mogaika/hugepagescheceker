<a href="https://hub.docker.com/r/mogaika/hugepageschecker" alt="Build">
    <img src="https://img.shields.io/docker/cloud/build/mogaika/hugepageschecker" /></a>

# hugepagescheceker
Allow to check your linux hugepages setup. Can be used inside docker/kubernetes.
Must work for any page size, but tested only on 2M pages.
Docker image: ```mogaika/hugepageschecker```

### Sample outputs
- Everything good
  ```
  INFO: Test passed for /dev/hugepages
  ```
- Permissions problems
  ```
  ERROR: Create hugepage file syscall error: permission denied
  ```
- Problems with your configuration ([k8s example](https://github.com/kubernetes/kubernetes/issues/71233#issuecomment-516061026))
  ```
  INFO: Trying to write page data
  [signal SIGBUS: bus error code=0x2 addr=0x7fcfa0200000 pc=0x4ad197]
  ```
- Hugepages not configured
  ```
  INFO: Mounts found: []
  ```

- Out of hugepages memory
  ```
  ERROR: Unable to map page: cannot allocate memory
  ```

### Kubernetes pod spec
```yaml
spec:
  containers:
  - name: hugepageschecker
    image: mogaika/hugepageschecker
    imagePullPolicy: Always
    command:
      - sh
      - -c
      - "/hugepageschecker && sleep 999999999"
    resources:
      limits:
        hugepages-2Mi: 2Mi
        memory: 128Mi
    volumeMounts:
    - mountPath: /hugepages
      name: hugepages
  volumes:
  - name: hugepages
    emptyDir:
      medium: HugePages
```
