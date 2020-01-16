<a href="https://img.shields.io/docker/cloud/build/mogaika/hugepageschecker" alt="Build">
    <img src="https://img.shields.io/docker/cloud/build/mogaika/hugepageschecker" /></a>

# hugepagescheceker
Tool allow to check your linux hugepages setup. Can be used inside docker/kubernetes

# example kubernetes pod
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
