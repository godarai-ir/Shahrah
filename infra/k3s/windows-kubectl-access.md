# Windows kubectl Access via Bastion

This runbook documents the working Windows and PowerShell access path for the Shahrah K3s cluster from a laptop.

## Topology

- bastion public SSH endpoint: `ubuntu@37.152.191.92`
- `node-1`: `10.10.0.21`
- `node-2`: `10.10.0.22`
- `node-3`: `10.10.0.23`
- private K3s API load balancer: HAProxy + Keepalived on `lb-1` and `lb-2`
- private K3s API endpoint: `10.10.0.6:6443`

Important:

- `kubectl` should not target `node-1` directly for normal laptop access
- the stable path is laptop -> bastion -> VIP `10.10.0.6:6443`
- local `kubectl` should talk to `https://127.0.0.1:16443`
- describe this path as HAProxy + Keepalived; the host-level edge is not Nginx
- the Keepalived VIP `10.10.0.6` was validated from bastion on 2026-07-16
- the Kubernetes API certificate includes `10.10.0.6` and `k3s-api.internal` in SAN
- direct tunnels to `10.10.0.7:6443` or `10.10.0.8:6443` are no longer valid; HAProxy listens on the VIP

## Local prerequisites

- Windows OpenSSH client available at `C:\Windows\System32\OpenSSH\ssh.exe`
- `kubectl` installed on the laptop
- an OpenSSH private key for the bastion
  - if you only have a PuTTY `.ppk`, export it as an OpenSSH private key first
- a kubeconfig context that points to the local forwarded port

## SSH config

Create `~/.ssh/config` with:

```sshconfig
Host bastion
    HostName 37.152.191.92
    User ubuntu
    Port 22
    IdentityFile ~/.ssh/ar-iran-vps-0
    IdentitiesOnly yes
    ServerAliveInterval 30
    ServerAliveCountMax 3
    StrictHostKeyChecking accept-new

Host node-1
    HostName 10.10.0.21
    User ubuntu
    Port 22
    IdentityFile ~/.ssh/ar-iran-vps-0
    IdentitiesOnly yes
    ProxyJump bastion
    ServerAliveInterval 30
    ServerAliveCountMax 3
    StrictHostKeyChecking accept-new

Host node-2
    HostName 10.10.0.22
    User ubuntu
    Port 22
    IdentityFile ~/.ssh/ar-iran-vps-0
    IdentitiesOnly yes
    ProxyJump bastion
    ServerAliveInterval 30
    ServerAliveCountMax 3
    StrictHostKeyChecking accept-new

Host node-3
    HostName 10.10.0.23
    User ubuntu
    Port 22
    IdentityFile ~/.ssh/ar-iran-vps-0
    IdentitiesOnly yes
    ProxyJump bastion
    ServerAliveInterval 30
    ServerAliveCountMax 3
    StrictHostKeyChecking accept-new

Host lb-1
    HostName 10.10.0.7
    User ubuntu
    Port 22
    IdentityFile ~/.ssh/ar-iran-vps-0
    IdentitiesOnly yes
    ProxyJump bastion
    ServerAliveInterval 30
    ServerAliveCountMax 3
    StrictHostKeyChecking accept-new

Host shahrah-k3s-api
    HostName 10.10.0.7
    User ubuntu
    Port 22
    IdentityFile ~/.ssh/ar-iran-vps-0
    IdentitiesOnly yes
    ProxyJump bastion
    ExitOnForwardFailure yes
    LocalForward 16443 10.10.0.6:6443
    ServerAliveInterval 30
    ServerAliveCountMax 3
    StrictHostKeyChecking accept-new
```

## Kubeconfig expectation

The production context should use the local forwarded endpoint:

```yaml
clusters:
- cluster:
    server: https://127.0.0.1:16443
  name: default

contexts:
- context:
    cluster: default
    user: default
  name: shahrah-production

current-context: shahrah-production
```

The CA and client certificate data must come from the live cluster admin kubeconfig on `node-1`.

## Start the tunnel

Foreground:

```powershell
ssh -N shahrah-k3s-api
```

Background:

```powershell
Start-Process -FilePath 'C:\Windows\System32\OpenSSH\ssh.exe' `
  -ArgumentList @('-N','shahrah-k3s-api') `
  -WindowStyle Hidden
```

## Verify SSH connectivity

```powershell
ssh bastion hostname
ssh node-1 hostname
ssh lb-1 hostname
```

Expected hostnames:

- `bastion`
- `node-1`
- `lb-1`

## Verify kubectl access

```powershell
kubectl config use-context shahrah-production
kubectl cluster-info
kubectl get nodes -o wide
kubectl auth whoami
```

Expected result:

- control plane reachable at `https://127.0.0.1:16443`
- all three control-plane nodes in `Ready`
- authenticated user is `system:admin`

## Failure modes and fixes

### 1. `connectex: No connection could be made because the target machine actively refused it`

Meaning:

- nothing is listening on local port `16443`

Checks:

```powershell
Test-NetConnection 127.0.0.1 -Port 16443
Get-NetTCPConnection -LocalPort 16443
```

Fix:

- start the SSH tunnel again with `ssh -N shahrah-k3s-api`

### 1.1. `wsarecv: An existing connection was forcibly closed by the remote host`

Meaning:

- the local kubeconfig points to `https://127.0.0.1:16443`, but the SSH tunnel is forwarding to an old or wrong Kubernetes API endpoint
- after the VIP cutover, `10.10.0.7:6443` and `10.10.0.8:6443` are not the operational API listeners

Fix:

- update the tunnel target to `10.10.0.6:6443`
- close any old process listening on local port `16443`
- start the tunnel again with `ssh -N shahrah-k3s-api`

### 2. `You must be logged in to the server` or `Unauthorized`

Meaning:

- the tunnel is fine
- the local kubeconfig client certificate is stale or belongs to a different cluster state

Confirmed real incident:

- local laptop had an older `system:admin` client certificate
- live cluster admin kubeconfig on `node-1` had newer cert material
- replacing only the client certificate and key fixed access immediately

Refresh procedure:

```powershell
ssh node-1 "sudo cat /etc/rancher/k3s/k3s.yaml"
```

Then:

1. back up the local kubeconfig
2. keep the cluster server as `https://127.0.0.1:16443`
3. replace the `default` user `client-certificate-data`
4. replace the `default` user `client-key-data`
5. keep the live `certificate-authority-data`

Recommended backup command:

```powershell
Copy-Item "$HOME\.kube\config" "$HOME\.kube\config.backup-$(Get-Date -Format yyyy-MM-dd)"
```

After refresh:

```powershell
kubectl cluster-info
kubectl get nodes -o wide
kubectl auth whoami
```

## Known-good cluster checks

These commands were confirmed from `node-1` during debugging:

```bash
sudo k3s kubectl get nodes -o wide
```

Expected current node IPs:

- `node-1`: `10.10.0.21`
- `node-2`: `10.10.0.22`
- `node-3`: `10.10.0.23`

## Notes

- the old PuTTY tunnel worked because it forwarded `L16443=10.10.0.7:6443`
- the current tunnel should forward `L16443=10.10.0.6:6443`
- forwarding to `node-1` SSH alone is not enough for `kubectl`
- use the private VIP endpoint for the normal access path
