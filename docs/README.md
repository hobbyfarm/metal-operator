# metal-operator

Launch and manage Equinix metal instances using K8S.

The project supports to crds:
* Instance
* ImportKeyPair

### Instance
The Instance type can be used to launch Equinix Metal Servers in your account.

Sample manifest is as follows:
```
apiVersion: equinix.cattle.io/v1alpha1
kind: Instance
metadata:
  name: instance-sample
spec:
  plan: c3.small.x86
  credentialSecret: equinix-metal
  facility:
    - sy4
  operatingSystem: "custom_ipxe"
  billingCycle: "hourly"
  ipxeScriptUrl: "https://raw.githubusercontent.com/ibrokethecloud/custom_pxe/master/shell.ipxe"
  projectsshKeys:
    - bf389b19-2fcf-4f02-9a6f-939e7482c821
```

*Note*: This example is using a custom pxe script which leaves the device in shell prompt.

### ImportKeyPair
The ImportKeyPair type can be used to create a KeyPair in Equinix Metal project using your custom public key.

Sample manifest is as follows:
```
apiVersion: equinix.cattle.io/v1alpha1
kind: ImportKeyPair
metadata:
  name: importkeypair-sample
spec:
  key: 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDF6c8PcOyIUSELn2RtiRx+9FIvw0siU7qpvkg94uLA4qscs3CUXIOjfQkt/3UdVf2KmM/YMbBeTVAoCu6PQ7lsFtXijvldVF+S11YYB2Fz8r9VYPYnpgTurxc3UcDtPC4OYUGJYlG5n7ddPOWAHjc6ZY5hcvilxLioFWHpMUGAKTeS6AWJ1fb6MW4iw/xXsb3evM4rEgTX4ya3uzW6HFOz5pI5yD3oN7LTTaRBI6aaUblLJ6HIfbyoL/IAs4SkF1qpqtcKkurkpmVHgShPc+WCtkofzhmS2UZK/a/ztAHkJ8ny1UNXodeCMF7FCQN10rZr2VUggb6i1B9RGHnK7BJv'
  secret: equinix-metal
```

For both custom types the secret is a k8s secret which contains the keys `PACKET_AUTH_TOKEN` and `PROJECT_ID`

Easiest way to generate one is follows:

```
kubectl create secret aws-secret --from-literal=PROJECT_ID="MYPROJECTID" --from-literal=PACKET_AUTH_TOKEN="MYTOKEN" -n metal-operator
```

To get started a helm chart is available [here.](./charts/metal-operator)

Quick installation:

```
kubectl create namepsace metal-operator
helm install metal-operator ./charts/metal-operator -n metal-operator
```