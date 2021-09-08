package metal

import (
	"fmt"
	"strings"

	equinixv1alpha1 "github.com/hobbyfarm/metal-operator/pkg/api/v1alpha1"
	"github.com/packethost/packngo"
)

func (m *MetalClient) createKeyPair(sshCreateRequest *packngo.SSHKeyCreateRequest) (keyPairID string, err error) {

	sshKey, _, err := m.SSHKeys.Create(sshCreateRequest)
	if err != nil {
		return keyPairID, err
	}

	keyPairID = sshKey.ID
	return keyPairID, err
}

func (m *MetalClient) DeleteKeyPair(importKeyPair *equinixv1alpha1.ImportKeyPair) (err error) {
	ok, err := m.importKeyPairExists(importKeyPair.Status.KeyPairID)
	if err != nil {
		return err
	}

	// key exists lets delete
	if ok {
		_, err = m.SSHKeys.Delete(importKeyPair.Status.KeyPairID)
		return err
	}

	// key no longer exists.. ignore this object
	return nil
}

func (m *MetalClient) CreateImportKeyPair(importKeyPair *equinixv1alpha1.ImportKeyPair) (status *equinixv1alpha1.ImportKeyPairStatus, err error) {
	status = importKeyPair.Status.DeepCopy()

	sshReq := m.generateSSHKeyRequest(importKeyPair)
	keyPairID, err := m.createKeyPair(sshReq)
	if err != nil {
		return status, err
	}
	status.KeyPairID = keyPairID
	status.Status = "created"
	return status, nil
}

func (m *MetalClient) importKeyPairExists(keyPairID string) (ok bool, err error) {
	sshKey, _, err := m.SSHKeys.Get(keyPairID, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not found") {
			return ok, nil
		} else {
			return ok, err
		}
	}

	if sshKey.ID == keyPairID {
		ok = true
	}
	return ok, nil
}

func (m *MetalClient) generateSSHKeyRequest(importKeyPair *equinixv1alpha1.ImportKeyPair) (sshReq *packngo.SSHKeyCreateRequest) {
	sshReq = &packngo.SSHKeyCreateRequest{}
	sshReq.ProjectID = m.ProjectID
	sshReq.Key = importKeyPair.Spec.Key
	sshReq.Label = fmt.Sprintf("%s-%s", importKeyPair.Name, importKeyPair.Namespace)

	return sshReq
}
