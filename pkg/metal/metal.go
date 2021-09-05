package metal

import (
	"context"
	"fmt"

	equinixv1alpha1 "github.com/hobbyfarm/metal-operator/pkg/api/v1alpha1"
	"github.com/packethost/packngo"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MetalClient struct {
	*packngo.Client
	ProjectID string
}

func NewClient(ctx context.Context, client client.Client, secret string, namespace string) (m *MetalClient, err error) {
	m = &MetalClient{}
	credSecret := &corev1.Secret{}

	err = client.Get(ctx, types.NamespacedName{Name: secret, Namespace: namespace}, credSecret)

	if err != nil {
		return nil, errors.Wrap(err, "error during credential secret lookup")
	}

	key, ok := credSecret.Data["PACKET_AUTH_TOKEN"]
	if !ok {
		return nil, fmt.Errorf("no key PACKET_AUTH_TOKEN found in secret %s", secret)
	}

	m.Client = packngo.NewClientWithAuth("packngo lib", string(key), nil)

	projectID, ok := credSecret.Data["PROJECT_ID"]
	if !ok {
		return nil, fmt.Errorf("no key PROJECT_ID specified in secret %s", secret)
	}

	m.ProjectID = string(projectID)
	return m, nil
}

func (m *MetalClient) CreateNewDevice(instance *equinixv1alpha1.Instance) (status *equinixv1alpha1.InstanceStatus, err error) {
	status = instance.Status.DeepCopy()
	dsr := m.generateDeviceCreationRequest(instance)
	device, _, err := m.Devices.Create(dsr)
	if err != nil {
		return status, errors.Wrap(err, "error during device creation")
	}

	status.InstanceID = device.ID
	status.Status = device.State
	return status, err
}

func (m *MetalClient) generateDeviceCreationRequest(instance *equinixv1alpha1.Instance) (dsr *packngo.DeviceCreateRequest) {
	dsr = &packngo.DeviceCreateRequest{
		Hostname:              fmt.Sprintf("%s-%s", instance.Name, instance.Namespace),
		Plan:                  instance.Spec.Plan,
		Facility:              instance.Spec.Facility,
		Metro:                 instance.Spec.Metro,
		ProjectID:             m.ProjectID,
		AlwaysPXE:             instance.Spec.AlwaysPXE,
		Tags:                  instance.Spec.Tags,
		Description:           instance.Spec.Description,
		PublicIPv4SubnetSize:  instance.Spec.PublicIPv4SubnetSize,
		HardwareReservationID: instance.Spec.HardwareReservationID,
		SpotInstance:          instance.Spec.SpotInstance,
		SpotPriceMax:          instance.Spec.SpotPriceMax,
		CustomData:            instance.Spec.CustomData,
		UserSSHKeys:           instance.Spec.UserSSHKeys,
		ProjectSSHKeys:        instance.Spec.ProjectSSHKeys,
		Features:              instance.Spec.Features,
		NoSSHKeys:             instance.Spec.NoSSHKeys,
		OS:                    instance.Spec.OS,
		BillingCycle:          instance.Spec.BillingCycle,
		IPXEScriptURL:         instance.Spec.IPXEScriptURL,
	}

	return dsr
}

func (m *MetalClient) CheckDeviceStatus(instance *equinixv1alpha1.Instance) (status *equinixv1alpha1.InstanceStatus, err error) {
	status = instance.Status.DeepCopy()
	deviceStatus, _, err := m.Devices.Get(instance.Status.InstanceID, nil)
	if err != nil {
		return status, err
	}

	if deviceStatus.State == "active" {
		status.Status = "active"
		status.PrivateIP = deviceStatus.GetNetworkInfo().PrivateIPv4
		status.PublicIP = deviceStatus.GetNetworkInfo().PublicIPv4
	}

	return status, nil
}

func (m *MetalClient) DeleteDevice(instance *equinixv1alpha1.Instance) (err error) {

	if instance.Status.Status != "deleted" {
		_, err = m.Devices.Delete(instance.Status.InstanceID, true)
		if err != nil {
			return err
		}
	}

	instance.Status.Status = "deleted"
	return nil
}
