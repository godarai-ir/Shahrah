# Argo CD Bootstrap Notes

This folder captures the chosen bootstrap and repository-access model for Shahrah.

## Chosen repository access method

For `shahrah-gitops`, the selected access method is:

- `SSH deploy key`

Why this was chosen:

- the GitOps repo uses GitHub SSH over port 443: `ssh://git@ssh.github.com:443/Delpak/shahrah-gitops.git`
- it avoids changing existing `Application` manifests to HTTPS
- it keeps Argo CD access read-only and scoped to one repository
- it is cleaner than reusing a personal SSH key

This means:

- create one dedicated SSH keypair only for Argo CD
- add the public key to GitHub as a read-only deploy key on `Delpak/shahrah-gitops`
- store the private key in an Argo CD repository secret in namespace `argocd`
- the private key used by Argo CD must be unencrypted and must not require a passphrase

Security note:

- do not commit the real private key into Git
- if a real private key was ever pasted into a tracked file, rotate that deploy key in GitHub and generate a new one
- Argo CD cannot unlock a passphrase-protected SSH private key during repository access

## Required outbound access

Argo CD needs outbound access for two separate things:

1. node-level image pulls
2. runtime Git access from Argo CD components such as `argocd-repo-server`

Current chosen egress workaround:

- VPN runs on `bastion`
- `squid` runs on `bastion` at `10.10.0.5:3128`
- K3s nodes use that proxy for K3s/containerd image pulls
- Argo CD deployments should also receive proxy environment variables if Git access needs the same egress path
- GitHub SSH port 22 is blocked from this network; repository access uses `ssh.github.com:443` with the same deploy key

## Bootstrap sequence

Run these in order:

1. install Argo CD in `argocd`
2. ensure Argo CD pods are healthy
3. create repo secret using `repo-shahrah-gitops.secret.yaml`
4. load GitHub SSH host key into `argocd-ssh-known-hosts-cm`
5. apply:
   - `projects/shahrah-project.yaml`
   - `bootstrap/app-of-apps.yaml`
6. verify that `shahrah-root` and child apps sync

## Files in this folder

- `repo-shahrah-gitops.secret.yaml`
  - template for the Argo CD repository secret
- `github-known-hosts.configmap.yaml`
  - template for GitHub SSH known hosts

Do not commit the real private key into Git.

If you edit the YAML templates by hand:

- every line under `ssh_known_hosts: |` must stay indented
- every line under `sshPrivateKey: |` must stay indented
- if even one key line starts at column 1, Kubernetes YAML parsing will fail

## Recommended apply method

Prefer generating the secret and known-hosts objects from the live files on the server instead of editing the YAML templates by hand.

Examples:

```bash
kubectl -n argocd create secret generic repo-shahrah-gitops \
  --from-literal=type=git \
  --from-literal=url=ssh://git@ssh.github.com:443/Delpak/shahrah-gitops.git \
  --from-literal=project=shahrah \
  --from-file=sshPrivateKey=/path/to/argocd_shahrah_gitops \
  --dry-run=client -o yaml | \
  kubectl label -f - argocd.argoproj.io/secret-type=repository --local -o yaml | \
  kubectl apply -f -
```

```bash
ssh-keyscan github.com > /tmp/github_known_hosts
kubectl -n argocd create configmap argocd-ssh-known-hosts-cm \
  --from-file=ssh_known_hosts=/tmp/github_known_hosts \
  --dry-run=client -o yaml | kubectl apply -f -
```
